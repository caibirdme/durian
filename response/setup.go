package response

import (
	"bytes"
	super "github.com/caibirdme/caddy-fasthttp/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"regexp"
	"strconv"
	"strings"
)

var (
	defaultStatusCode  = 200
	defaultContentType = "text/html; charset=utf-8"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveResponse, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

type RespConfig struct {
	Path        string
	Pattern     string
	Code        int
	Body        string
	ContentType string
	Headers     []super.KVTuple
}

func setup(c *caddy.Controller) error {
	cfg, err := parseCfg(c)
	if err != nil {
		return err
	}
	var re *regexp.Regexp
	if cfg.Pattern != "" {
		re, err = regexp.Compile(cfg.Path)
		if err != nil {
			return err
		}
	}
	prefixBytes := []byte(cfg.Path)
	super.GetConfig(c).AddMiddleware(func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if re != nil {
				if re.Match(ctx.Path()) {
					outputDirectly(ctx, cfg)
				} else {
					next(ctx)
				}
			} else {
				if bytes.HasPrefix(ctx.Path(), prefixBytes) {
					outputDirectly(ctx, cfg)
				} else {
					next(ctx)
				}
			}
		}
	})
	return nil
}

func outputDirectly(ctx *fasthttp.RequestCtx, cfg *RespConfig) {
	ctx.SetStatusCode(cfg.Code)
	ctx.SetBodyString(cfg.Body)
	ctx.SetContentType(cfg.ContentType)
	for _, item := range cfg.Headers {
		ctx.Response.Header.Set(item.K, item.V)
	}
}

func parseCfg(c *caddy.Controller) (*RespConfig, error) {
	// skip response keyword
	c.Next()
	cfg := RespConfig{
		Code:        defaultStatusCode,
		ContentType: defaultContentType,
	}
	for c.NextBlock() {
		kind := c.Val()
		switch strings.ToLower(kind) {
		case "code":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			code, err := strconv.Atoi(c.Val())
			if err != nil {
				return nil, c.Err(err.Error())
			}
			cfg.Code = code
		case "content_type":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			cfg.ContentType = c.Val()
		case "body":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			cfg.Body = strings.Trim(c.Val(), `"`)
		case "header":
			var k, v string
			if !c.Args(&k, &v) {
				return nil, c.ArgErr()
			}
			cfg.Headers = append(cfg.Headers, super.KVTuple{K: k, V: v})
		case "pattern":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			if cfg.Path != "" {
				return nil, c.Errf("pattern and path is exclusive")
			}
			cfg.Pattern = c.Val()
		case "path":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			if cfg.Pattern != "" {
				return nil, c.Errf("pattern and path is exclusive")
			}
			cfg.Path = c.Val()
		}
	}
	if cfg.Path == "" && cfg.Pattern == "" {
		return nil, c.Err("must specify path or pattern")
	}
	return &cfg, nil
}
