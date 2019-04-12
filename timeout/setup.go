package timeout

import (
	"time"

	super "github.com/caibirdme/caddy-fasthttp"
	"github.com/mholt/caddy"
)

const (
	pluginName = "timeout"
)

func init() {
	caddy.RegisterPlugin(pluginName, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setupTimeouts,
	})
}

func setupTimeouts(c *caddy.Controller) error {
	cfg := super.GetConfig(c)
	if cfg == nil {
		return c.Errf("[%s] couldn't find %s's config", pluginName, c.Key)
	}

	for c.Next() {
		var hasOptionalBlock bool
		for c.NextBlock() {
			hasOptionalBlock = true

			// ensure the kind of timeout is recognized
			kind := c.Val()
			if kind != "read" && kind != "write" {
				return c.Errf("unknown timeout '%s': must be read, header, write, or idle", kind)
			}

			// parse the timeout duration
			if !c.NextArg() {
				return c.ArgErr()
			}
			if c.NextArg() {
				// only one value permitted
				return c.ArgErr()
			}
			var dur time.Duration
			if c.Val() != "none" {
				var err error
				dur, err = time.ParseDuration(c.Val())
				if err != nil {
					return c.Errf("%v", err)
				}
				if dur < 0 {
					return c.Err("non-negative duration required for timeout value")
				}
			}

			// set this timeout's duration
			switch kind {
			case "read":
				cfg.ReadTimeout = dur
			case "write":
				cfg.WriteTimeout = dur
			}
		}
		if !hasOptionalBlock {
			// set all timeouts to the same value

			if !c.NextArg() {
				return c.ArgErr()
			}
			if c.NextArg() {
				// only one value permitted
				return c.ArgErr()
			}
			val := c.Val()

			if val == "none" {
				cfg.ReadTimeout = 0
				cfg.WriteTimeout = 0
			} else {
				dur, err := time.ParseDuration(val)
				if err != nil {
					return c.Errf("unknown timeout duration: %v", err)
				}
				if dur < 0 {
					return c.Err("non-negative duration required for timeout value")
				}
				cfg.ReadTimeout = dur
				cfg.WriteTimeout = dur
			}
		}
	}

	return nil
}
