package header

import (
	super "github.com/caibirdme/durian/server"
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
	h := NewHeaderSetter(cfg.Headers)
	super.GetConfig(c).AddMiddleware(func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if cfg.location.Match(ctx.Path()) {
				if err := h.Set(ctx); err != nil {
					// todo: log
				}
			}
			next(ctx)
		}
	})
	return nil
}

type HeaderConfig struct {
	location super.LocationMatcher
	Headers  []super.KVTuple
}

// header path k v
//
// header path {
// 	k1 v1
//  k2 v2
// }
func parseHeader(c *caddy.Controller) (*HeaderConfig, error) {
	c.Next()

	firstLine := c.RemainingArgs()
	if len(firstLine) == 0 {
		return nil, c.ArgErr()
	}
	var cfg HeaderConfig
	var err error
	cfg.location, err = super.NewLocationMatcher(firstLine)
	if err != nil {
		return nil, err
	}

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
