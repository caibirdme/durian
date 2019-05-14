package upstream

import (
	super "github.com/caibirdme/durian/server"
	"github.com/mholt/caddy"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveUpstream, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	u, err := parseUpstream(c)
	if err != nil {
		return err
	}
	upstreamManager := c.Get(super.UpstreamKey)
	if upstreamManager != nil {
		m, ok := upstreamManager.(map[string]super.Upstream)
		if !ok {
			return errors.New("[Bug] upstreamManager isn't map[string]Upstream")
		}
		m[u.Name] = *u
		c.Set(super.UpstreamKey, m)
	}
	m := make(map[string]super.Upstream)
	m[u.Name] = *u
	c.Set(super.UpstreamKey, m)
	return nil
}

func parseUpstream(c *caddy.Controller) (*super.Upstream, error) {
	c.Next()

	if !c.NextArg() {
		return nil, c.ArgErr()
	}
	u := super.Upstream{}
	u.Name = c.Val()

	for c.NextBlock() {
		str_list := []string{c.Val()}
		if remain := c.RemainingArgs(); len(remain) > 0 {
			str_list = append(str_list, remain...)
		}
		if len(str_list) == 0 {
			return nil, c.ArgErr()
		}
		b := super.Backend{}
		if strings.HasPrefix(str_list[0], "unix:") {
			b.Network = "unix"
			b.Addr = str_list[0][5:]
		} else {
			b.Network = "tcp"
			b.Addr = str_list[0]
		}
		for i := 1; i < len(str_list); i++ {
			kv := strings.Split(str_list[i], "=")
			if len(kv) == 1 {
				switch kv[0] {
				case "backup":
				default:
					return nil, c.Errf("%s should be in the form of k=v", str_list[i])
				}
			} else {
				k, v := kv[0], kv[1]
				switch k {
				case "weight":
					weight, err := strconv.Atoi(v)
					if err != nil {
						return nil, c.Errf("value of weight must be int but %s", v)
					}
					b.Weight = weight
				default:
					//todo: warn log
				}
			}
		}
		u.Backends = append(u.Backends, b)
	}
	return &u, nil
}
