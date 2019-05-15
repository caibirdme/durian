package reverse_proxy

import (
	"strconv"
	"strings"
	"time"

	super "github.com/caibirdme/durian/server"
	"github.com/mholt/caddy"
)

const (
	pluginName = "proxy"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveProxy, caddy.Plugin{
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

type ProxyConfig struct {
	location         super.LocationMatcher
	AddressList      []string
	UpstreamHeader   []super.KVTuple
	DownstreamHeader []super.KVTuple
	Timeout          time.Duration
	MaxConn          int
}

var (
	defaultTimeout = 2 * time.Second
)

func parseProxy(c *caddy.Controller) (*ProxyConfig, error) {
	c.Next()
	cfg := ProxyConfig{Timeout: defaultTimeout}
	firstLine := c.RemainingArgs()
	location, err := super.NewLocationMatcher(firstLine)
	if err != nil {
		return nil, err
	}
	cfg.location = location
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
	case "max_conn":
		if !c.NextArg() {
			return c.Err("need path value")
		}
		max_conn, err := strconv.Atoi(c.Val())
		if err != nil {
			return c.Err(err.Error())
		}
		cfg.MaxConn = max_conn
	default:
		return c.Errf("[%s] illegal directive %s", pluginName, kind)
	}
	return nil
}

func parseKVTuple(c *caddy.Controller, cfg *[]super.KVTuple) error {
	var k, v string
	if !c.Args(&k, &v) {
		return c.ArgErr()
	}
	*cfg = append(*cfg, super.KVTuple{K: k, V: v})
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
