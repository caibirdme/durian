package root

import (
	"bytes"
	super "github.com/caibirdme/caddy-fasthttp/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"strings"
)

func init() {
	caddy.RegisterPlugin("root", caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	// skip root
	c.Next()

	// prefix
	if !c.NextArg() {
		return c.ArgErr()
	}
	prefix := c.Val()

	// root dir
	if !c.NextArg() {
		return c.ArgErr()
	}
	fs := &fasthttp.FS{
		Root: c.Val(),
	}
	if prefix != "/" {
		fs.PathRewrite = fasthttp.NewPathPrefixStripper(len(strings.TrimRight(prefix, "/")))
	}
	prefixBytes := []byte(prefix)
	process := fs.NewRequestHandler()
	super.GetConfig(c).AddMiddleware(func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if !bytes.HasPrefix(ctx.Path(), prefixBytes) {
				next(ctx)
			} else {
				process(ctx)
			}
		}
	})

	return nil
}
