package reverse_proxy

import (
	"log"
	"strings"
	"time"

	super "github.com/caibirdme/caddy-fasthttp/server"
	"github.com/mholt/caddy"
)

const (
	pluginName = "proxy"
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
	log.Println(cfg)
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
	// proxy
	if !c.NextArg() {
		return nil, c.ArgErr()
	}
	// load pattern
	if !c.NextArg() {
		return nil, c.ArgErr()
	}
	cfg := ProxyConfig{}
	cfg.Pattern = c.Val()
	for c.NextBlock() {
		kind := c.Val()
		err := parseKind(kind, &cfg, c)
		if nil != err {
			return nil, err
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

// due to bug of Dispender, this is a workaround
func parseUpstream(c *caddy.Controller, cfg *ProxyConfig) error {
	// {
	if !c.Next() {
		return c.ArgErr()
	}
	for c.Next() {
		address := c.Val()
		if address == "}" {
			break
		}
		cfg.AddressList = append(cfg.AddressList, address)
	}
	if len(cfg.AddressList) == 0 {
		return c.Errf("[%s] upstream should contain at least one address", pluginName)
	}
	return nil
}
