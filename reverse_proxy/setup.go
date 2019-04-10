package reverse_proxy

import (
	super "github.com/caibirdme/caddy-fasthttp"
	"github.com/mholt/caddy"
)

const (
	pluginName = "reverse_proxy"
)

func init() {
	caddy.RegisterPlugin(pluginName, caddy.Plugin{
		ServerType:super.FastHTTPServerType,
		Action: setup,
	})
}

func setup(c *caddy.Controller) error {
	super.GetConfig(c).AddMiddleware(ProxyHandler)
	return nil
}
