package status

import (
	"bytes"
	super "github.com/caibirdme/durian/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"strconv"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveStatus, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	cfg, err := parseStatus(c)
	if err != nil {
		return err
	}
	super.GetConfig(c).AddMiddleware(func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			for _, basePath := range cfg.Paths {
				if bytes.HasPrefix(ctx.Path(), []byte(basePath)) {
					ctx.SetStatusCode(cfg.Code)
					break
				}
			}
			next(ctx)
		}
	})
	return nil
}

type StatusConfig struct {
	Code  int
	Paths []string
}

func parseStatus(c *caddy.Controller) (*StatusConfig, error) {
	// skip status keyword
	c.Next()
	if !c.NextArg() {
		return nil, c.ArgErr()
	}
	code, err := strconv.Atoi(c.Val())
	if err != nil {
		return nil, c.Err("statusCode must be a valid number defined by RFC")
	}
	cfg := StatusConfig{Code: code}
	hasBlock := false
	for c.NextBlock() {
		hasBlock = true
		cfg.Paths = append(cfg.Paths, c.Val())
	}
	// no block just path
	if !hasBlock {
		if !c.NextArg() {
			return nil, c.Err("must specify a path")
		} else {
			cfg.Paths = append(cfg.Paths, c.Val())
		}
	}
	if len(cfg.Paths) == 0 {
		return nil, c.Err("must specify at least one path")
	}
	return &cfg, nil
}
