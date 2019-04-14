package timeout

import (
	super "github.com/caibirdme/caddy-fasthttp/server"
	"github.com/mholt/caddy"
	"strings"
	"time"
)

const (
	pluginName = "timeout"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveTimeout, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setupTimeouts,
	})
}

func setupTimeouts(c *caddy.Controller) error {
	cfg := super.GetConfig(c)
	if cfg == nil {
		return c.Errf("[%s] couldn't find %s's config", pluginName, c.Key)
	}
	c.Next()
	for c.NextBlock() {
		kind := c.Val()
		switch strings.ToLower(kind) {
		case "keep_alive":
			d, err := time.ParseDuration(c.Val())
			if err != nil {
				return c.Err(err.Error())
			}
			cfg.MaxKeepaliveDuration = d
		}
	}
	return nil
}
