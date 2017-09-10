## fcgipoold

This package is a [fcgi](https://fast-cgi.github.io/)-multiplexer that maintains a pool of fcgi worker processes while exposing a single fcgi interface. This kinda similar to apache's [fcgid module](https://httpd.apache.org/mod_fcgid/mod/mod_fcgid.html) or [php-fpm](https://php-fpm.org/) and can be used e.g. with nginx. This is the missing link between a webserver and processes that implement the fcgi protocol.

### Getting started

While being in the root dir of this repository run `go build cmd/pool/poold.go` and `go build cmd/client/client.go` to have the binaries in this directory and then run the poold process with `./poold -c ./client -n 10 -s 8000 -r 50`. You should have a webserver running that forwards the incoming http requests via fcgi to a different process. See below for a nginx configuration. The next step assumes that you have nginx setup with the configuration below.

To test that everything is setup properly and runs fine run `curl http://localhost` and `curl -X POST http://localhost`. That should output `Hello from FCGI child` and `Hello from FCGI child via POST` respectively.

example nginx configuration
```
server {
    listen       80;
    server_name  localhost;
    location / {
            fastcgi_param SERVER_PROTOCOL $server_protocol;
            fastcgi_param REQUEST_METHOD $request_method;
            fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
            fastcgi_pass 127.0.0.1:9000;
    }
}
```

### Limitiations & Status
This is a experimental POC and at this point in time only simple HTTP GET and POST Requests are supported. Feel free to contribute.