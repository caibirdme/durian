package gzip

import (
	"strconv"
	"strings"

	super "github.com/caibirdme/caddy-fasthttp/server"
	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveGzip, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	cfg, err := parseGzip(c)
	if err != nil {
		return err
	}
	if cfg == nilConfig {
		return nil
	}
	if cfg.Level == 0 {
		return nil
	}
	srvCfg := super.GetConfig(c)
	if srvCfg == nil {
		return nil
	}
	srvCfg.Gzip.Open = true
	srvCfg.Gzip.Level = cfg.Level
	return nil
}

type GzipConfig struct {
	Level int
}

var nilConfig = GzipConfig{}

const (
	defaultGzipLevel = 6
)

func parseGzip(c *caddy.Controller) (GzipConfig, error) {
	// skip gzip keyword
	c.Next()
	// if only gzip and no block, use defaultGzipLevel
	cfg := GzipConfig{Level: defaultGzipLevel}

	for c.NextBlock() {
		kind := c.Val()
		switch strings.ToLower(kind) {
		case "level":
			if !c.NextArg() {
				return nilConfig, c.ArgErr()
			}
			level, err := strconv.Atoi(c.Val())
			if err != nil {
				return nilConfig, c.Err(err.Error())
			}
			cfg.Level = level
		}
	}
	return cfg, nil
}
