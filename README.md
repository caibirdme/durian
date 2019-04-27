# Durian

[中文](./docs/zh-CN.md)

Durian is a generic purpose web server, like apache,nginx and caddy(Durian is based on caddy).

And, Durian is also a modular and pluggable framework.

See examples to have a quick start.

## Generic Web Server
Durian itself is a powerful and high performance web server. You can use it like nginx:
```
:8051 {

    # if url matches pattern, then response directly
    response {
        pattern /foo/\d+/.*
        body "{\"name\": \"caibirdme\", \"age\": 25, \"some\":[12,3,4]}"
        content_type Application/json
    }
    
    # reverse proxy
    proxy {
        pattern /bar/.*
        upstream {
            10.10.10.1
            10.10.10.2
        }
    }

    # configure log
    log {
        access_path ./access.log
        err_path ./error.log
        # if no more words, this means using default log format
        # for more infomartion about the log format, see doc for log module
        format
    }
}
```
Config file above can sets up a server listening on port 8051 and does response and reverse proxy as you wish.
each entry in access.log is a json, such as:
```json
{"remote_addr":"127.0.0.1:50926","host":"localhost:8051","method":"GET","request_uri":"/user/deen","status":200,"start_time":"2019-04-27T19:56:51.811+0800","process_time":"19.337µs","bytes_sent":85,"user_agent":"curl/7.58.0","response_body":"Hello deen\n"}
```
```json
{"remote_addr":"127.0.0.1:50928","host":"localhost:8051","method":"GET","request_uri":"/foo/123/bar","status":200,"start_time":"2019-04-27T19:57:06.733+0800","process_time":"12.34µs","bytes_sent":155,"user_agent":"curl/7.58.0","response_body":"{\"name\": \"caibirdme\", \"age\": 25, \"some\":[12,3,4]}"}
```
### Advantage of Durian
* easy to configure
* high performance(at least 5 times faster than Go's net/http)
* zero-downtime
    * graceful shutdown
    * graceful restart
    * graceful upgrade
* cross platform
* could work with supervisor
* highly extensible(really!!)

### Extend Durian
Like OpenResty, you can use Lua to extend the function of Nginx.
You can extend Durian with pure go due to Durian's modular design, with only no more than 10 lines, which is really very easy.

Below is a full and runnable example:
```go
package main

import (
	"fmt"
	"github.com/buaazp/fasthttprouter"
	"github.com/caibirdme/durian"
	"github.com/caibirdme/durian/router"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"log"
)

func init() {
	// trap signals and enable graceful restart
	caddy.TrapSignals()
}

func main() {
	caddy.AppName = "haha"
	caddy.AppVersion = "0.0.1"
	// Register your own logic here
	router.RegisterPlugin(handler)

	// read config
	input, err := durian.ReadConfig("./Caddyfile")

	// start the server
	instance, err := caddy.Start(input)
	if err != nil {
		log.Fatal(err)
	}

	// wait for all servers to stop
	instance.Wait()
}

// your logic entry, the main function for your plugin
func handler(cfg router.RouterConfig) (fasthttp.RequestHandler, error) {
	/*
	# in your Caddyfile
	router {
	    config /home/my/app.toml
	}
	 */
	
	// cfg.CfgPath == "/home/my/app.toml"
	_ = cfg.CfgPath
	

	// whatever router or framework(based on fasthttp)
	r := fasthttprouter.New()
	// configure route rules
	r.GET("/user/:name", getUserName)

	return r.Handler, nil
}

func getUserName(ctx *fasthttp.RequestCtx) {
	ctx.WriteString(fmt.Sprintf("Hello %v\n", ctx.UserValue("name")))
}
```

See, It's really easy, and now you have all the abilities Durian provides.


This project is still in progress, blow are the current supported directives:

```
:8051 {
    # set timeout for server
    timeout {
        keep_alive 5m
        read 10s
        write 10s
    }
    
    gzip {
        level 6
    }
        
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
    
    response {
        pattern /hello/\w+
        body "{\"name\": \"caibirdme\", \"age\": 25, \"some\":[12,3,4]}"
        content_type Application/json
    }
    
    log {
        access_path ./access.log
        err_path ./error.log
        format {
            remote_addr
            host
            method
            request_uri
            status
            start_time 
            process_time 
            bytes_sent
            user_agent
            response_body 
        }
    }
}

:8012 {
    concurrency 1000  # The maximum number of concurrent connections the server may serve
    root / /home/my/sitedir # root simply specifies the root of the site
}
```

## Directives

### response
if path matched, return directly. This is always helpful when you want to set up a mock server

#### syntax
```
response {
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

### NotFound
not_found is used to specify the action when url mismatch
#### syntax
```
not_found {
    subdirectives
    #...
}
```
#### subdirectives
* `file string`: specify a file to send to client
* `body string`: specify a string to send to client, default "not found"
* `code int`: specify the response status code, default 404
* `content_type string`: set content type, default "text/html; charset=utf-8"

**note**: can't set file and body at the same time
#### example
```
not_found {
    file /var/www/site/404.html
}
```
### log
log related config, each entry is in json format
#### syntax
```
log {
    subdirective
    #...
}
```
#### subdirecitve
* `access_path string`: specify the path of access log
    * `access_path /foo/bar`: access log will be stored in /foo/bar/access.log
    * `access_path /foo/bar/customize.log`: access log will be stored in customize.log
* `err_path string`: specify the path of error log
* `format {entries...}`: specify access log content
    * now
    * bytes_sent
    * body_bytes_sent
    * connection_requests
    * request_time
    * request_length
    * status
    * user_agent
    * remote_addr
    * request_uri
    * query_string
    * request_body
    * request_header
    * method
    * response_body
    * response_header
    * referer
#### example
```
log {
    access_path /var/site/test.log
    err_path /var/site/err.log
    format {
        now
        remote_addr
        bytes_sent
        method
        request_uri
        status
        user_agent
    }
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

#### durian
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
From the benchmark above, durian is more than 4 times faster than caddy.
More benchmarks are on the way...