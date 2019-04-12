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
    root . # root simply specifies the root of the site
}
```

## Plan

- [ ] rewrite
- [ ] circuit breaker for reverse proxy
- [ ] dynamic upstream for reverse proxy(watch specified file)
- [ ] request_id
- [ ] rate limit
- [ ] many other directives caddy supported yet...