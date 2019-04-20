package timeout

import (
	"strings"
	"time"

	super "github.com/caibirdme/durian/server"
	"github.com/mholt/caddy"
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
		if !c.NextArg() {
			return c.ArgErr()
		}
		d, err := time.ParseDuration(c.Val())
		if err != nil {
			return c.Err(err.Error())
		}
		switch strings.ToLower(kind) {
		case "keep_alive":
			cfg.MaxKeepaliveDuration = d
		case "read":
			cfg.ReadTimeout = d
		case "write":
			cfg.WriteTimeout = d
		}
	}
	return nil
}
