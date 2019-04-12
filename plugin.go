package caddy_fasthttp

import (
	// plug in server
	_ "github.com/caibirdme/caddy-fasthttp/server"

	// plug in directives
	_ "github.com/caibirdme/caddy-fasthttp/reverse_proxy"
	_ "github.com/caibirdme/caddy-fasthttp/timeout"
)
