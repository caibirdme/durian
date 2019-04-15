# caddy-fasthttp

Make caddy run 10 times faster, thx to fasthttp's high performance

See examples for help.

This project is in progress, blow are the current supported directives:

```
:8051 {
    # set timeout for server
    timeout {
        keep_alive 30s
    }
    
    gzip {
        level 6
    }
    
    status 400 /abc
    
    # /api/asd/hello_123 -> /foo/hello_123/other/asd
    rewrite /api/(\w+)/(.*) {
        to /foo/{2}/other/{1}
    }

    # reverse proxy
    proxy {
        pattern /foo/(\w)/(.*) 
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
proxy {
    subdirecitves
    #...
}
```
#### Subdirectives
* `pattern string`: url pattern to match
* `path string`: url prefix to match
* `upstream block`: specify upstream address, one address per line. The address is the form of `ip:port`
* `timeout duration`: timeout for waiting upstream response
* `header_upstream string string`: header added to upstream
* `header_downstream string string`: header added to downstream
* `max_conn int`: max connections to keep for upstream

note: pattern and path is exclusively required
#### examples
```
proxy {
    pattern /foo/bar/.* 
    upstream {
        10.10.18.3:8000
        10.10.19.4:7000
    }
    timeout 300ms # 1s 2min...
    header_upstream X-Foo foo
    header_upstream X-Other ttt
    header_downstream X-Bar "hello client"
    header_downstream Access-Control-Allow-Origin *
    max_conn 1000
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
* `keep_alive duration`: set keepalive duration
* `read duration`: set readTimeout, the time spend to read data from the connection.
* `write duration`: set writeTimeout, writeTimeout should largeEqual than `readTimeout+processTime`

**note**: For detailed explanations about timeout, see: [here](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)

#### example
```
timeout {
    keep_alive 30s
    read 1s
    write 2s
}
```

### header
set extra header to request
#### syntax
```
header path key val

header path {
    key1 val1
    key2 val2
    #...
}
```
#### example
```
header / {
    X-Server caddy_fast
}
```

### Gzip
enable gzip if request contains gzip related header
#### syntax
```
gzip #default level 6

gzip {
    subdirectives
    #...
}
```
#### subdirectives
* `level int`: set gzip level
#### example
```
gzip {
    level 7
}
``` 

## Plan

- [x] rewrite
- [ ] circuit breaker for reverse proxy
- [ ] dynamic upstream for reverse proxy(watch specified file)
- [ ] request_id
- [ ] rate limit
- [ ] many other directives caddy supported yet...

## Benchmark

### machine
```
CPU: i7-8700 12cores 3.2GHz
MEM: 16G
OS: Linux caibirdme-MS-7B53 4.15.0-47-generic #50-Ubuntu SMP Wed Mar 13 10:44:52 UTC 2019 x86_64 x86_64 x86_64 GNU/Linux
```

### upstream server sleep 50ms

#### caddy-fasthttp
##### Caddyfile
```
:8051
proxy {
    path /foo
    upstream {
        localhost:9001
        localhost:9002
    }
    max_conn 1000
}
```
##### test command
`wrk -c950 -t2 -d20s http://localhost:8051/foo/bar`
##### result
```
Running 20s test @ http://localhost:8051/foo/bar
  2 threads and 950 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    50.59ms  635.91us  70.82ms   95.06%
    Req/Sec     9.41k   355.56     9.60k    97.50%
  374795 requests in 20.09s, 63.62MB read
Requests/sec:  18651.61
Transfer/sec:      3.17MB
```
#### caddy
##### caddyfile
```
:8051

proxy /foo localhost:9001 localhost:9002
```
##### test command
`wrk -c950 -t2 -d20s http://localhost:8051/foo/bar`
##### result
```
Running 20s test @ http://localhost:8051/foo/bar
  2 threads and 950 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   270.50ms  364.35ms   2.00s    86.32%
    Req/Sec     2.34k     2.72k    9.60k    84.78%
  89122 requests in 20.07s, 14.89MB read
  Socket errors: connect 0, read 0, write 0, timeout 1369
  Non-2xx or 3xx responses: 1178
Requests/sec:   4440.60
Transfer/sec:    759.75KB
```

#### conclusion
From the benchmark above, caddy-fasthttp is more than 4 times faster than caddy.
More benchmarks are on the way...