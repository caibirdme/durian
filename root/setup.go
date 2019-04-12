package root

import (
	super "github.com/caibirdme/caddy-fasthttp/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
)

func init() {
	caddy.RegisterPlugin("root", caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	c.Next()
	if !c.NextArg() {
		return c.ArgErr()
	}
	fs := &fasthttp.FS{
		Root: c.Val(),
	}

	super.GetConfig(c).AddMiddleware(func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return fs.NewRequestHandler()
	})

	return nil
}
