package header

import (
	"bytes"
	super "github.com/caibirdme/caddy-fasthttp/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveHeader, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	cfg, err := parseHeader(c)
	if err != nil {
		return err
	}
	pathPrefixBytes := []byte(cfg.Path)
	super.GetConfig(c).AddMiddleware(func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if bytes.HasPrefix(ctx.Path(), pathPrefixBytes) {
				for _, pair := range cfg.Headers {
					ctx.Request.Header.Set(pair.K, pair.V)
				}
			}
			next(ctx)
		}
	})
	return nil
}

type HeaderConfig struct {
	Path    string
	Headers []super.KVTuple
}

// header path k v
//
// header path {
// 	k1 v1
//  k2 v2
// }
func parseHeader(c *caddy.Controller) (*HeaderConfig, error) {
	c.Next()

	if !c.NextArg() {
		return nil, c.ArgErr()
	}
	cfg := HeaderConfig{Path: c.Val()}

	hasBlock := false
	for c.NextBlock() {
		var k, v string
		if !c.Args(&k, &v) {
			return nil, c.ArgErr()
		}
		cfg.Headers = append(cfg.Headers, super.KVTuple{K: k, V: v})
		hasBlock = true
	}
	if !hasBlock {
		var k, v string
		if !c.Args(&k, &v) {
			return nil, c.ArgErr()
		}
		cfg.Headers = append(cfg.Headers, super.KVTuple{K: k, V: v})
	}
	return &cfg, nil
}
