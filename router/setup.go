package router

import (
	"sync"
	super "github.com/caibirdme/durian/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"strings"
)

type RegisterFunc func(cfg RouterConfig) (fasthttp.RequestHandler, error)

var once sync.Once
var userDefinedHandler RegisterFunc

func RegisterPlugin(fn RegisterFunc) {
	once.Do(func() {
		userDefinedHandler = fn
		Init()
	})
}

func Init() {
	caddy.RegisterPlugin(super.DirectiveRouter, caddy.Plugin{
		ServerType:super.FastHTTPServerType,
		Action:setup,
	})
}

func setup(c *caddy.Controller) error {
	cfg, err := parseConfig(c)
	if err != nil {
		return err
	}
	handler, err := userDefinedHandler(cfg)
	if err != nil {
		return err
	}
	super.GetConfig(c).AddNamedMiddleware(super.RouterMiddlewareName, func (next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return handler
	})
	return nil
}

type RouterConfig struct {
	CfgPath string
}

func parseConfig(c *caddy.Controller) (RouterConfig, error) {
	c.Next()
	cfg := RouterConfig{}
	for c.NextBlock() {
		kind := c.Val()
		switch strings.ToLower(kind) {
		case "config":
			if !c.NextArg() {
				return cfg, c.ArgErr()
			}
			cfg.CfgPath = c.Val()
		}
	}
	return cfg, nil
}
