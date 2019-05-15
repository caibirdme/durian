package response

import (
	"strconv"
	"strings"

	"github.com/caibirdme/durian/header"
	super "github.com/caibirdme/durian/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
)

var (
	defaultStatusCode  = 200
	defaultContentType = "text/html; charset=utf-8"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveResponse, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

type RespConfig struct {
	location    super.LocationMatcher
	Code        int
	Body        string
	ContentType string
	Headers     []super.KVTuple
}

func setup(c *caddy.Controller) error {
	cfg, err := parseCfg(c)
	if err != nil {
		return err
	}
	h := header.NewHeaderSetter(cfg.Headers)
	super.GetConfig(c).AddMiddleware(func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if cfg.location.Match(ctx.Path()) {
				if err := h.Set(ctx); err != nil {
					// todo: log
				}
				outputDirectly(ctx, cfg)
			} else {
				next(ctx)
			}
		}
	})
	return nil
}

func outputDirectly(ctx *fasthttp.RequestCtx, cfg *RespConfig) {
	ctx.SetStatusCode(cfg.Code)
	ctx.SetBodyString(cfg.Body)
	ctx.SetContentType(cfg.ContentType)
}

func parseCfg(c *caddy.Controller) (*RespConfig, error) {
	// skip response keyword
	c.Next()
	cfg := RespConfig{
		Code:        defaultStatusCode,
		ContentType: defaultContentType,
	}
	firstLie := c.RemainingArgs()
	location, err := super.NewLocationMatcher(firstLie)
	if err != nil {
		return nil, err
	}
	cfg.location = location
	for c.NextBlock() {
		kind := c.Val()
		switch strings.ToLower(kind) {
		case "code":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			code, err := strconv.Atoi(c.Val())
			if err != nil {
				return nil, c.Err(err.Error())
			}
			cfg.Code = code
		case "content_type":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			cfg.ContentType = c.Val()
		case "body":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			cfg.Body = strings.Trim(c.Val(), `"`)
		case "header":
			var k, v string
			if !c.Args(&k, &v) {
				return nil, c.ArgErr()
			}
			cfg.Headers = append(cfg.Headers, super.KVTuple{K: k, V: v})
		}
	}
	return &cfg, nil
}
