package caddy_fasthttp

import (
	// plug in server
	_ "github.com/caibirdme/caddy-fasthttp/server"

	// plug in directives
	_ "github.com/caibirdme/caddy-fasthttp/gzip"
	_ "github.com/caibirdme/caddy-fasthttp/header"
	_ "github.com/caibirdme/caddy-fasthttp/response"
	_ "github.com/caibirdme/caddy-fasthttp/reverse_proxy"
	_ "github.com/caibirdme/caddy-fasthttp/rewrite"
	_ "github.com/caibirdme/caddy-fasthttp/root"
	_ "github.com/caibirdme/caddy-fasthttp/status"
	_ "github.com/caibirdme/caddy-fasthttp/timeout"
)
