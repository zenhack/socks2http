package main

import (
	"flag"
	"golang.org/x/net/proxy"
	"io"
	"log"
	"net"
	"net/http"
)

var (
	laddr = flag.String("laddr", ":8000", "Address to listen on")
	raddr = flag.String("raddr", ":1080", "Socks proxy address to connect to.")
)

var (
	socks5proxy proxy.Dialer
	client      *http.Client
)

func newClient(dialer proxy.Dialer) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(net, addr string) (net.Conn, error) {
				return socks5proxy.Dial(net, addr)
			},
		},
	}
}

func main() {
	flag.Parse()
	var err error
	socks5proxy, err = proxy.SOCKS5("tcp", *raddr, nil, proxy.Direct)
	if err != nil {
		panic(err)
	}
	client = newClient(socks5proxy)
	hndl := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "CONNECT" {
			serverConn, err := socks5proxy.Dial("tcp", req.Host)
			log.Println("Got server connection")
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error() + "\n"))
				return
			}
			hijacker, ok := w.(http.Hijacker)
			if !ok {
				serverConn.Close()
				w.WriteHeader(500)
				w.Write([]byte("Failed cast to Hijacker\n"))
				return
			}
			w.WriteHeader(200)
			_, bio, err := hijacker.Hijack()
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error() + "\n"))
				serverConn.Close()
				return
			}
			go io.Copy(serverConn, bio)
			go io.Copy(bio, serverConn)
		} else {
			// Server-Only field; we get an error fi we pass this to `client.Do`.
			req.RequestURI = ""

			resp, err := client.Do(req)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error() + "\n"))
				return
			}
			hdr := w.Header()
			for k, v := range resp.Header {
				hdr[k] = v
			}
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
		}
	})
	log.Fatal(http.ListenAndServe(*laddr, hndl))
}
