package reverse_proxy

import (
	"strings"
	"time"

	super "github.com/caibirdme/caddy-fasthttp"
	"github.com/mholt/caddy"
)

const (
	pluginName = "reverse_proxy"
)

func init() {
	caddy.RegisterPlugin(pluginName, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	cfg, err := parseProxy(c)
	if nil != err {
		return err
	}
	p, err := NewProxy(*cfg)
	if nil != err {
		return err
	}
	super.GetConfig(c).AddMiddleware(p.Handle)
	return nil
}

type KVTuple struct {
	K, V string
}

type ProxyConfig struct {
	Pattern          string
	AddressList      []string
	UpstreamHeader   []KVTuple
	DownstreamHeader []KVTuple
	Timeout          time.Duration
}

func parseProxy(c *caddy.Controller) (*ProxyConfig, error) {
	cfg := ProxyConfig{}
	for c.Next() {
		if !c.NextArg() {
			return nil, c.Errf("[%s] url pattern is needed", pluginName)
		}
		cfg.Pattern = c.Val()
		for c.NextBlock() {
			kind := c.Val()
			err := parseKind(kind, &cfg, c)
			if nil != err {
				return nil, err
			}
		}
	}
	return &cfg, nil
}

func parseKind(kind string, cfg *ProxyConfig, c *caddy.Controller) error {
	switch strings.ToLower(kind) {
	case "timeout":
		if !c.NextArg() {
			return c.Errf("[%s] timeout value isn't set", pluginName)
		}
		val := c.Val()
		d, err := time.ParseDuration(val)
		if nil != err {
			return err
		}
		cfg.Timeout = d
	case "header_upstream":
		err := parseKVTuple(c, &cfg.UpstreamHeader)
		if nil != err {
			return err
		}
	case "header_downstream":
		err := parseKVTuple(c, &cfg.DownstreamHeader)
		if nil != err {
			return err
		}
	case "upstream":
		err := parseUpstream(c, cfg)
		if nil != err {
			return err
		}
	default:
		return c.Errf("[%s] illegal directive %s", pluginName, kind)
	}
	return nil
}

func parseKVTuple(c *caddy.Controller, cfg *[]KVTuple) error {
	var k, v string
	if !c.Args(&k, &v) {
		return c.ArgErr()
	}
	*cfg = append(*cfg, KVTuple{K: k, V: v})
	return nil
}

func parseUpstream(c *caddy.Controller, cfg *ProxyConfig) error {
	for c.NextBlock() {
		address := c.Val()
		cfg.AddressList = append(cfg.AddressList, address)
	}
	if len(cfg.AddressList) == 0 {
		return c.Errf("[%s] upstream should contain at least one address", pluginName)
	}
	return nil
}
