package fastcgi

import (
	"fmt"
	super "github.com/caibirdme/durian/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"regexp"
	"time"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveFastCgi, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

type Config struct {
	Debug       bool
	KeepConn    bool
	ReadTimeout time.Duration
	SendTimeout time.Duration
	Upstream    super.Upstream
}

func setup(c *caddy.Controller) error {
	fmt.Println("setup fcgi")
	cfg, rule, err := parseFcgiCfg(c)
	if err != nil {
		return err
	}
	super.GetConfig(c).AddMiddleware(func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		var srv *Handler
		srv, err = NewHandler(rule, cfg, next)
		if err != nil {
			return next
		}
		return srv.Serve
	})
	return err
}

func parseFcgiCfg(c *caddy.Controller) (*Config, *Rule, error) {
	c.Next()

	if !c.NextArg() {
		return nil, nil, c.ArgErr()
	}

	rule := Rule{
		Params: make(map[string]string),
	}
	cfg := Config{}

	pattern, err := regexp.Compile(c.Val())
	if err != nil {
		return nil, nil, err
	}
	rule.Pattern = pattern

	for c.NextBlock() {
		list := getLine(c)
		if len(list) == 0 {
			return nil, nil, c.ArgErr()
		}
		switch list[0] {
		case "root":
			if len(list) > 1 {
				rule.Root = list[1]
			}
		case "index":
			if len(list) > 1 {
				rule.Index = list[1]
			}
		case "split_path_info":
			if len(list) > 1 {
				re, err := regexp.Compile(list[1])
				if err != nil {
					return nil, nil, err
				}
				// test correctness
				testMatch := re.FindStringSubmatch("/foo/bar/abc.php/qq/tt/index.php")
				if len(testMatch) != 3 || testMatch[1] != "/foo/bar/abc.php" && testMatch[2] != "/qq/tt/index.php" {
					return nil, nil, c.Err("regexp of split_path_info isn't correct")
				}
				rule.SplitPathInfo = re
			}
		case "script_filename_prefix":
			if len(list) > 1 {
				rule.FilenamePrefix = list[1]
			}
		case "catch_stderr":
			if len(list) > 1 {
				rule.CatchStderr = list[1]
			}
		case "server_software":
			if len(list) > 1 {
				rule.ServerSoftware = list[1]
			}
		case "server_name":
			if len(list) > 1 {
				rule.ServerName = list[1]
			}
		case "debug":
			cfg.Debug = true
		case "keep_conn":
			cfg.KeepConn = true
		case "read_timeout":
			if len(list) > 1 {
				d, err := time.ParseDuration(list[1])
				if err != nil {
					return nil, nil, c.Errf("read_timeout isn't a duration: %s", err)
				}
				cfg.ReadTimeout = d
			}
		case "send_timeout":
			if len(list) > 1 {
				d, err := time.ParseDuration(list[1])
				if err != nil {
					return nil, nil, c.Errf("send_timeout isn't a duration: %s", err)
				}
				cfg.SendTimeout = d
			}
		case "upstream":
			if len(list) > 1 {
				name := list[1]
				if v := c.Get(super.UpstreamKey); v != nil {
					m := v.(map[string]super.Upstream)
					if u, ok := m[name]; ok {
						cfg.Upstream = u
					} else {
						return nil, nil, c.Errf("invalid upstream name %s", name)
					}
				} else {
					return nil, nil, c.Errf("invalid upstream name %s", name)
				}
			}
		case "fcgi_param":
			if len(list) > 2 {
				rule.Params[list[1]] = list[2]
			}
		}
	}
	if rule.SplitPathInfo == nil {
		rule.SplitPathInfo = regexp.MustCompile(`^(.+?\.php)(/.*)$`)
	}
	if rule.Root == "" {
		if root := c.Get(super.DocRootKey); root != nil {
			rule.Root = root.(string)
		}
	}
	if rule.ServerName == "" {
		rule.ServerName = caddy.AppName
	}
	if rule.ServerSoftware == "" {
		rule.ServerSoftware = super.DurianName
	}
	if rule.Index == "" {
		rule.Index = "index.php"
	}
	fmt.Printf("cfg: %+v\nrule:%+v\n", cfg, rule)
	return &cfg, &rule, nil
}

func getLine(c *caddy.Controller) []string {
	list := []string{c.Val()}
	remain := c.RemainingArgs()
	if len(remain) > 0 {
		list = append(list, remain...)
	}
	return list
}
