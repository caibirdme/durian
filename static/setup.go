package static

import (
	"bytes"
	super "github.com/caibirdme/durian/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"strings"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveStatic, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

var (
	rootPath = []byte("/")
)

func setup(c *caddy.Controller) error {
	cfg := StaticConfig{}
	err := parseStatic(c, &cfg)
	if err != nil {
		return err
	}
	srvCfg := super.GetConfig(c)
	if cfg.Root == "" {
		cfg.Root = srvCfg.Root
	}
	fs := &fasthttp.FS{
		Root:       cfg.Root,
		Compress:   cfg.Compress,
		IndexNames: cfg.Index,
	}
	if cfg.prefix != "/" {
		stripper := fasthttp.NewPathPrefixStripper(len(strings.TrimRight(cfg.prefix, "/")))
		fs.PathRewrite = func(ctx *fasthttp.RequestCtx) []byte {
			newPath := stripper(ctx)
			if len(newPath) == 0 {
				return rootPath
			} else {
				return newPath
			}
		}
	}
	prefixBytes := []byte(cfg.prefix)
	process := fs.NewRequestHandler()
	srvCfg.AddMiddleware(func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if !bytes.HasPrefix(ctx.Path(), prefixBytes) {
				next(ctx)
			} else {
				process(ctx)
			}
		}
	})

	return nil
}

type StaticConfig struct {
	prefix   string
	Root     string
	Index    []string
	Compress bool
}

func parseStatic(c *caddy.Controller, cfg *StaticConfig) error {
	// skip root
	c.Next()

	// prefix
	if !c.NextArg() {
		return c.ArgErr()
	}
	cfg.prefix = c.Val()
	hasBlock := false
	for c.NextBlock() {
		hasBlock = true
		kind := c.Val()
		switch strings.ToLower(kind) {
		case "root":
			if !c.NextArg() {
				return c.ArgErr()
			}
			cfg.Root = c.Val()
		case "compress":
			cfg.Compress = true
		case "index":
			for c.NextArg() {
				cfg.Index = append(cfg.Index, c.Val())
			}
		}
	}
	if !hasBlock {
		// root dir
		if !c.NextArg() {
			return c.ArgErr()
		}
		cfg.Root = c.Val()
	}
	return nil
}
