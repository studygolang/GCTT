Translating by polaris1119!

# HTTP(S) Proxy in Golang in less than 100 lines of code

![](https://cdn-images-1.medium.com/max/1600/1*9FR0CCERIPLgv5SDAFlpOQ.jpeg)

The goal is to implement a [proxy server](https://en.wikipedia.org/wiki/Proxy_server) for HTTP and HTTPS. Handling of HTTP is a matter of parsing request, passing such request to destination server, reading response and passing it back to the client. All we need for that is built-in HTTP server and client ([net/http](https://golang.org/pkg/net/http/)). HTTPS is different as it’ll use technique called HTTP CONNECT tunneling. First client sends request using HTTP CONNECT method to set up the tunnel between the client and destination server. When such tunnel consisting of two TCP connections is ready, client starts regular TLS handshake with destination server to establish secure connection and later send requests and receive responses.

## Certificates

Our proxy will be an HTTPS server (when `—-proto https` will be used) so we need certificate and private key. For the purpose of this post let’s use self-signed certificate. To generate one use such script:

```console
#!/usr/bin/env bash
case `uname -s` in
    Linux*)     sslConfig=/etc/ssl/openssl.cnf;;
    Darwin*)    sslConfig=/System/Library/OpenSSL/openssl.cnf;;
esac
openssl req \
    -newkey rsa:2048 \
    -x509 \
    -nodes \
    -keyout server.key \
    -new \
    -out server.pem \
    -subj /CN=localhost \
    -reqexts SAN \
    -extensions SAN \
    -config <(cat $sslConfig \
        <(printf '[SAN]\nsubjectAltName=DNS:localhost')) \
    -sha256 \
    -days 3650
```
It’s required to convince your OS to trust such certificate. In OS X it can be done with Keychain Access — https://tosbourn.com/getting-os-x-to-trust-self-signed-ssl-certificates/.

## HTTP

To support HTTP we’ll use [built-in HTTP server and client](https://golang.org/pkg/net/http/). The role of proxy is to handle HTTP request, pass such request to destination server and send response back to the client.

```
   +------+        +-----+        +-----------+
   |client|        |proxy|        |destination|
   +------+        +-----+        +-----------+
1          --Req-->       
2                         --Req-->
3                         <--Res--
4          <--Res--
```

## HTTP CONNECT tunneling

Suppose client wants to use either HTTPS or WebSockets in order to talk to server. Client is aware of using proxy. Simple HTTP request / response flow cannot be used since client needs to f.ex. establish secure connection with server (HTTPS) or wants to use other protocol over TCP connection (WebSockets). Technique which works is to use HTTP [CONNECT](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT) method. It tells the proxy server to establish TCP connection with destination server and when done to proxy the TCP stream to and from the client. This way proxy server won’t terminate SSL but will simply pass data between client and destination server so these two parties can establish secure connection

```
    +------+            +-----+                   +-----------+
    |client|            |proxy|                   |destination| 
    +------+            +-----+                   +-----------+
1           --CONNECT-->       
2                              <--TCP handshake-->
3           <--------------Tunnel---------------->
```

## Implementation

```go
package main
import (
    "crypto/tls"
    "flag"
    "io"
    "log"
    "net"
    "net/http"
    "time"
)
func handleTunneling(w http.ResponseWriter, r *http.Request) {
    dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
    hijacker, ok := w.(http.Hijacker)
    if !ok {
        http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
        return
    }
    client_conn, _, err := hijacker.Hijack()
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
    }
    go transfer(dest_conn, client_conn)
    go transfer(client_conn, dest_conn)
}
func transfer(destination io.WriteCloser, source io.ReadCloser) {
    defer destination.Close()
    defer source.Close()
    io.Copy(destination, source)
}
func handleHTTP(w http.ResponseWriter, req *http.Request) {
    resp, err := http.DefaultTransport.RoundTrip(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    defer resp.Body.Close()
    copyHeader(w.Header(), resp.Header)
    w.WriteHeader(resp.StatusCode)
    io.Copy(w, resp.Body)
}
func copyHeader(dst, src http.Header) {
    for k, vv := range src {
        for _, v := range vv {
            dst.Add(k, v)
        }
    }
}
func main() {
    var pemPath string
    flag.StringVar(&pemPath, "pem", "server.pem", "path to pem file")
    var keyPath string
    flag.StringVar(&keyPath, "key", "server.key", "path to key file")
    var proto string
    flag.StringVar(&proto, "proto", "https", "Proxy protocol (http or https)")
    flag.Parse()
    if proto != "http" && proto != "https" {
        log.Fatal("Protocol must be either http or https")
    }
    server := &http.Server{
        Addr: ":8888",
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Method == http.MethodConnect {
                handleTunneling(w, r)
            } else {
                handleHTTP(w, r)
            }
        }),
        // Disable HTTP/2.
        TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
    }
    if proto == "http" {
        log.Fatal(server.ListenAndServe())
    } else {
        log.Fatal(server.ListenAndServeTLS(pemPath, keyPath))
    }
}
```

> Presented code is not a production-grade solution. It lacks f.ex. handling [hop-by-hop headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers#hbh), setting up timeouts while copying data between two connections or the ones exposed by net/http — more on this in [“The complete guide to Go net/http timeouts”](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/).

Our server while getting request will take one of two paths: handling HTTP or handling HTTP CONNECT tunneling. This is done with:

```go
http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodConnect {
        handleTunneling(w, r)
    } else {
        handleHTTP(w, r)
    }
})
```

Function to handle HTTP — handleHTTP is self-explanatory so let’s focus on handling tunneling. The first part of handleTunneling is about setting connection to destination server:

```go
dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
if err != nil {
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
}
w.WriteHeader(http.StatusOK)
```

Next we’ve a part to hijack connection maintained by HTTP server:

```go
hijacker, ok := w.(http.Hijacker)
if !ok {
    http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
    return
}
client_conn, _, err := hijacker.Hijack()
if err != nil {
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
}
```

[Hijacker interface](https://golang.org/pkg/net/http/#Hijacker) allows to take over the connection. After that the caller is responsible to manage such connection (HTTP library won’t do it anymore).

Once we’ve two TCP connections (client→proxy, proxy→destination server) we need to set tunnel up:

```go
go transfer(dest_conn, client_conn)
go transfer(client_conn, dest_conn)
```

In two goroutines data is copied in two directions: from the client to the destination server and backward.

## Testing

To test our proxy you can use f.ex. Chrome:

```go
> Chrome --proxy-server=https://localhost:8888
```

or [Curl](https://github.com/curl/curl):

```
> curl -Lv --proxy https://localhost:8888 --proxy-cacert server.pem https://google.com
```

> curl needs to be built with HTTPS-proxy support (introduced in 7.52.0).

## HTTP/2

In our server HTTP/2 support has been deliberately removed because then hijacking is not possible. More on this in [#14797](https://github.com/golang/go/issues/14797#issuecomment-196103814).


Click ❤ below to help others discover this story. Please follow me if you want to get updates about new posts or boost work on future stories.

----------------

via: https://medium.com/@mlowicki/http-s-proxy-in-golang-in-less-than-100-lines-of-code-6a51c2f2c38c

作者：[Michał Łowicki](https://medium.com/@mlowicki)
译者：[译者ID](https://github.com/译者ID)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出
