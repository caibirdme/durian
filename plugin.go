package caddy_fasthttp

import (
	"time"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyfile"
	"github.com/valyala/fasthttp"
)

const (
	defaultConcurrent  = 3000
	defaultReadTimeout = time.Second
)

const (
	FastHTTPServerType = "fasthttp"
)

func init() {
	caddy.RegisterServerType(FastHTTPServerType, caddy.ServerType{
		Directives: func() []string { return nil },
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

// ServerConfig stores the configuration for fasthttp.Server
type ServerConfig struct {
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
	middlewares                   []Middleware
}

func (cfg *ServerConfig) AddMiddleware(m Middleware) {
	cfg.middlewares = append(cfg.middlewares, m)
}

func (cfg *ServerConfig) makeServer() *fasthttp.Server {
	srv := &fasthttp.Server{
		Handler: compileMiddlewareEndWithNotFound(cfg.middlewares),
	}
	if cfg.ReadTimeout != 0 {
		srv.ReadTimeout = cfg.ReadTimeout
	}
	if cfg.WriteTimeout != 0 {
		srv.WriteTimeout = cfg.WriteTimeout
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
	return cfg, nil
}

func (c *fastContext) MakeServers() ([]caddy.Server, error) {
	return []caddy.Server{NewFastServer(c.cfg[0])}, nil
}

var directives = []string{
	"proxy",
	"header",
	"timeout",
}
