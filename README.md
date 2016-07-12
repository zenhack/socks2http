Simple tool to plumb http proxy requests through a socks5 proxy.

NOTE: this is a little rough. Known issues:

* We use most of the default settings for `net/http.Client`, so things
  like following redirects on the server-side will happen, which is
  probably not The Right Thing.

Also, I just haven't tested this super carefully. Buyer beware; you get
what you pay for.

Building:

    go build

Usage:

    socks2http -laddr <http-proxy-address> -raddr <socks-proxy-address>

Then point http proxy users at `<http-proxy-address>`

# License

MIT
