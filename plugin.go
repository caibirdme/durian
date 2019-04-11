package caddy_fasthttp

import (
	"time"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyfile"
	"github.com/valyala/fasthttp"
)

const (
	defaultConcurrent = 3000
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
	cfg  ServerConfig
}

func newContext(inst *caddy.Instance) caddy.Context {
	return &fastContext{
		inst: inst,
	}
}

func GetConfig(c *caddy.Controller) *ServerConfig {
	f := c.Context().(*fastContext)
	return &f.cfg
}

// serverConfig stores the configuration for fasthttp.Server
type ServerConfig struct {
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
	return &fasthttp.Server{
		Handler: compileMiddlewareEndWithNotFound(cfg.middlewares),
	}
}

func (c *fastContext) InspectServerBlocks(path string, sblocks []caddyfile.ServerBlock) ([]caddyfile.ServerBlock, error) {
	return sblocks, nil
}

func (c *fastContext) MakeServers() ([]caddy.Server, error) {
	return []caddy.Server{NewFastServer(c.cfg)}, nil
}
