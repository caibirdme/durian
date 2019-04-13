# caddy-fasthttp

Make caddy run 10 times faster, thx to fasthttp's high performance

See examples for help.

This project is in progress, blow are the current supported directives:

```
:8051 {
    # set timeout for server
    timeout {
        #Maximum duration for reading the full request (including body)
        # This also limits the maximum duration for idle keep-alive connections
        read 1s
        # Maximum duration for writing the full response (including body)
        write 1s
    }
    
    status 400 /abc
    
    # /api/asd/hello_123 -> /foo/hello_123/other/asd
    rewrite /api/(\w+)/(.*) {
        to /foo/{2}/other/{1}
    }

    # reverse proxy
    proxy /foo/(\w)/(.*) {
        upstream {
            localhost:8776
            localhost:8775
        }
        timeout 300ms
        header_upstream X-Foo ffoo
        header_upstream X-Bar barbar
        header_downstream X-Baz bazbaz
    }
}

:8012 {
    concurrency 1000  # The maximum number of concurrent connections the server may serve
    root / . # root simply specifies the root of the site
}
```

## Directives

### response
if path matched, return directly. This is always helpful when you want to set up a mock server

#### syntax
```
proxy {
    subdirectives
    #...
}
```
#### Subdirectives
* `path string`: path prefix to match
* `pattern string`: path pattern to match
* `code int`: status code, default 200
* `content_type string`: Content-Type header, default "text/html; charset=utf-8"
* `body string`: body to return(must add `"` if there're spaces in body)
* `header string string`: headers added to response

Note: path and pattern is exclusively required

#### examples
```
:8080 {
    response {
        pattern /foo/\d+/.*
        body "{\"name\": \"caibirdme\", \"age\":25}"
        content_type Application/json
    }
}
```
match /foo/123/whatever /foo/0/
```
:8080 {
    response {
        path /foo
        body "{\"name\": \"caibirdme\", \"age\":25}"
        content_type Application/json
    }
}
```
match all requests prefixed with /foo, such as /foo/1 /foo/bar/baz ...

### proxy
proxy requests to upstreams

#### syntax
```
proxy pattern {
    subdirecitves
    #...
}
```
#### Subdirectives
* `upstream block`: specify upstream address, one address per line. The address is the form of `ip:port`
* `timeout duration`: timeout for waiting upstream response
* `header_upstream string string`: header added to upstream
* `header_downstream string string`: header added to downstream
#### examples
```
proxy /foo/bar/.* {
    upstream {
        10.10.18.3:8000
        10.10.19.4:7000
    }
    timeout 300ms # 1s 2min...
    header_upstream X-Foo foo
    header_upstream X-Other ttt
    header_downstream X-Bar "hello client"
    header_downstream Access-Control-Allow-Origin *
}
```
Randomly reverse proxy request /foo/bar/xxx to 10.10.18.3:8000 or 10.10.19.4:7000

### root
set a directory as the root of a static file server

#### syntax
```
root url_prefix directory_path

root url_prefix {
    subdirectives
    #...
}
```
#### Subdirectives
* `dir string`: root directory path
* `compress`: cache compressed file to reduce usage of CPU(need write access to dir)
* `index string...`: specify index file
#### examples
```
root /download {
    dir /var/www/static
    compress
    index index.html index.htm index.json
}
```

### rewrite
rewrite url

#### syntax
````
rewrite pattern {
    to string
}
````
#### subdirectives
* `to string`: target pattern to rewrite, use {num} as placeholder
#### example
```
rewrite /foo/(\w+)/(\d+)/(.*) {
    to /{2}/bar/{1}/tee/{3}
}
```
This will rewrite `/foo/some/123/hello/world` to `/123/bar/some/tee/hello/world`

### timeout
set timeout for all requests
#### syntax
```
timeout {
    subdirectives
    #...
}
```
#### subdirectives
* `read duration`: set readTimeout
* `write duration`: set writeTimeout
#### example
```
timeout {
    read 1s
    write 800ms
}
```

## Plan

- [x] rewrite
- [ ] circuit breaker for reverse proxy
- [ ] dynamic upstream for reverse proxy(watch specified file)
- [ ] request_id
- [ ] rate limit
- [ ] many other directives caddy supported yet...