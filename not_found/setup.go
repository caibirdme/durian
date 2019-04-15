package not_found

import (
	super "github.com/caibirdme/caddy-fasthttp/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"os"
	"strconv"
	"strings"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveNotFound, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

const (
	defaultStatusCode  = fasthttp.StatusNotFound
	defaultBody        = "not found"
	defaultContentType = "text/html; charset=utf-8"
)

type NotFoundConfig struct {
	StatusCode  int
	File        string
	ContentType string
	Body        string
}

func setup(c *caddy.Controller) error {
	cfg, err := parseConfig(c)
	if err != nil {
		return err
	}
	if cfg.File != "" {
		fileInfo, err := os.Lstat(cfg.File)
		if err != nil {
			return c.Errf("%s path err: %s", cfg.File, err)
		}
		if fileInfo.IsDir() {
			return c.Errf("%s is a directory", cfg.File)
		}
	}
	srvCfg := super.GetConfig(c)
	if srvCfg == nil {
		panic("[BUG] server config can't be nil")
	}
	srvCfg.NotFound = super.NotFoundConfig{
		StatusCode:  cfg.StatusCode,
		File:        cfg.File,
		ContentType: cfg.ContentType,
		Body:        cfg.Body,
	}
	return nil
}

func parseConfig(c *caddy.Controller) (*NotFoundConfig, error) {
	c.Next()

	cfg := NotFoundConfig{StatusCode: defaultStatusCode, Body: defaultBody, ContentType: defaultContentType}
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
			cfg.StatusCode = code
		case "content_type":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			cfg.ContentType = c.Val()
		case "body":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			cfg.Body = c.Val()
		case "file":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			cfg.File = c.Val()
		}
	}
	if cfg.File != "" && cfg.Body != "" {
		return nil, c.Err("cannot specify file and body at the same time")
	}
	return &cfg, nil
}
