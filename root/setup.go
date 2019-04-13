package root

import (
	"bytes"
	super "github.com/caibirdme/caddy-fasthttp/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"strings"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveRoot, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

var (
	rootPath = []byte("/")
)

func setup(c *caddy.Controller) error {
	cfg := RootConfig{}
	err := parseRoot(c, &cfg)
	if err != nil {
		return err
	}
	fs := &fasthttp.FS{
		Root: cfg.Root,
		Compress:cfg.Compress,
		IndexNames:cfg.Index,
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
	super.GetConfig(c).AddMiddleware(func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
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

type RootConfig struct {
	prefix string
	Root string
	Index []string
	Compress bool
}

func parseRoot(c *caddy.Controller, cfg *RootConfig) error {
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
		case "dir":
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
