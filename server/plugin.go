package server

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyfile"
	"github.com/valyala/fasthttp"
)

const (
	FastHTTPServerType  = "fasthttp"
	RequestIDHeaderName = "rid"
)

func init() {
	caddy.RegisterServerType(FastHTTPServerType, caddy.ServerType{
		Directives: func() []string { return directives },
		DefaultInput: func() caddy.Input {
			return caddy.CaddyfileInput{
				ServerTypeName: FastHTTPServerType,
			}
		},
		NewContext: newContext,
	})
}

type fastContext struct {
	inst *caddy.Instance
	cfg  []ServerConfig
}

func newContext(inst *caddy.Instance) caddy.Context {
	return &fastContext{
		inst: inst,
	}
}

func GetConfig(c *caddy.Controller) *ServerConfig {
	f := c.Context().(*fastContext)
	for i := 0; i < len(f.cfg); i++ {
		if f.cfg[i].Addr == c.Key {
			return &f.cfg[i]
		}
	}
	return nil
}

type GzipConfig struct {
	Open  bool
	Level int
}

// ServerConfig stores the configuration for fasthttp.Server
type ServerConfig struct {
	Root                          string
	Addr                          string
	Name                          string
	Concurrency                   int
	DisableKeepalive              bool
	ReadBufferSize                int
	WriteBufferSize               int
	ReadTimeout                   time.Duration
	WriteTimeout                  time.Duration
	MaxConnsPerIP                 int
	MaxRequestsPerConn            int
	MaxKeepaliveDuration          time.Duration
	TCPKeepalive                  bool
	TCPKeepalivePeriod            time.Duration
	MaxRequestBodySize            int
	DisableHeaderNamesNormalizing bool
	NoDefaultServerHeader         bool
	NoDefaultContentType          bool
	Gzip                          GzipConfig
	NotFound                      NotFoundConfig
	middlewares                   []Middleware
	namedMiddleware               map[string]Middleware
	RequestIDName                 string
}

type NotFoundConfig struct {
	StatusCode  int
	File        string
	ContentType string
	Body        string
}

func (cfg *ServerConfig) AddMiddleware(m Middleware) {
	cfg.middlewares = append(cfg.middlewares, m)
}

func (cfg *ServerConfig) AddNamedMiddleware(name string, m Middleware) {
	if cfg.namedMiddleware == nil {
		cfg.namedMiddleware = make(map[string]Middleware)
	}
	cfg.namedMiddleware[name] = m
}

const (
	LogMiddlewareName    = "log"
	UUIDMiddlewareName   = "uuid"
	RouterMiddlewareName = "router"
)

func (cfg *ServerConfig) makeServer() *fasthttp.Server {
	var handler fasthttp.RequestHandler
	// mount user defined middleware
	if cfg.namedMiddleware != nil {
		if selfRouter, ok := cfg.namedMiddleware[RouterMiddlewareName]; ok {
			handler = selfRouter(notFoundHandler)
		}
	}
	// if there's not user defined middleware, just use not found as the final handler
	if handler == nil {
		handler = compileMiddleware(cfg.middlewares, notFoundHandler)
	} else {
		handler = compileMiddleware(cfg.middlewares, handler)
	}
	// mount uuid at the very beginning
	if cfg.namedMiddleware != nil {
		if m, ok := cfg.namedMiddleware[UUIDMiddlewareName]; ok {
			handler = m(handler)
		}
	}
	// mount the log as the outermost middleware
	if cfg.namedMiddleware != nil {
		if m, ok := cfg.namedMiddleware[LogMiddlewareName]; ok {
			handler = m(handler)
		}
	}
	srv := &fasthttp.Server{
		Handler: handler,
	}
	if d := cfg.MaxKeepaliveDuration; d != 0 {
		srv.MaxKeepaliveDuration = d
	}
	if d := cfg.ReadTimeout; d != 0 {
		srv.ReadTimeout = d
	}
	if d := cfg.WriteTimeout; d != 0 {
		srv.WriteTimeout = d
	}
	if cfg.Concurrency != 0 {
		srv.Concurrency = cfg.Concurrency
	}
	if cfg.DisableKeepalive {
		srv.DisableKeepalive = true
	}
	return srv
}

func (c *fastContext) InspectServerBlocks(path string, sblocks []caddyfile.ServerBlock) ([]caddyfile.ServerBlock, error) {
	for _, sblock := range sblocks {
		cfg, err := c.parseConfig(sblock)
		if nil != err {
			return sblocks, err
		}
		c.cfg = append(c.cfg, cfg)
	}
	return sblocks, nil
}

func (c *fastContext) parseConfig(sblock caddyfile.ServerBlock) (ServerConfig, error) {
	cfg := ServerConfig{Addr: sblock.Keys[0]}
	for key, vals := range sblock.Tokens {
		switch strings.ToLower(key) {
		case "concurrency":
			if len(vals) > 0 {
				val := vals[0].Text
				num, err := strconv.Atoi(val)
				if nil != err {
					return cfg, fmt.Errorf("%+v, Err: %s", vals, err)
				}
				cfg.Concurrency = num
			}
		case "keepalive":
			if len(vals) > 0 {
				val := vals[0].Text
				ok, err := strconv.ParseBool(val)
				if nil != err {
					return cfg, fmt.Errorf("%+v, Err: %s", vals, err)
				}
				// ok equals true means EnableKeepalive
				cfg.DisableKeepalive = !ok
			}
		case "root":
			if len(vals) > 0 {
				cfg.Root = vals[0].Text
			}
		}
	}
	if cfg.Root == "" {
		curDir, err := os.Getwd()
		if err != nil {
			return cfg, err
		}
		cfg.Root = curDir
	}
	return cfg, nil
}

func (c *fastContext) MakeServers() ([]caddy.Server, error) {
	var servers []caddy.Server
	for _, cfg := range c.cfg {
		servers = append(servers, NewFastServer(cfg))
	}
	return servers, nil
}

var directives = []string{
	DirectiveLog,
	DirectiveUpstream,
	DirectiveFastCgi,
	DirectiveGzip,
	DirectiveProxy,
	DirectiveHeader,
	DirectiveTimeout,
	DirectiveStatic,
	DirectiveRewrite,
	DirectiveStatus,
	DirectiveResponse,
	DirectiveNotFound,
	DirectiveRouter,
}

const (
	DirectiveProxy    = "proxy"
	DirectiveHeader   = "header"
	DirectiveTimeout  = "timeout"
	DirectiveStatic   = "static"
	DirectiveRewrite  = "rewrite"
	DirectiveStatus   = "status"
	DirectiveResponse = "response"
	DirectiveGzip     = "gzip"
	DirectiveNotFound = "not_found"
	DirectiveLog      = "log"
	DirectiveRouter   = "router"
	DirectiveFastCgi  = "fastcgi"
	DirectiveUpstream = "upstream"
)
