package caddy_fasthttp

import (
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyfile"
	"github.com/valyala/fasthttp"
	"time"
)

const (
	defaultConcurrent = 3000
)

func init() {
	caddy.RegisterServerType("fasthttp", caddy.ServerType{
		Directives: func() []string {return nil},
		DefaultInput: func() caddy.Input {return nil},
		NewContext: newContext,
	})
}

type fastContext struct {
	inst *caddy.Instance
	cfg serverConfig
}

func newContext(inst *caddy.Instance) caddy.Context {
	return &fastContext{
		inst: inst,
	}
}

// serverConfig stores the configuration for fasthttp.Server
type serverConfig struct {
	Name string
	Concurrency int
	DisableKeepalive bool
	ReadBufferSize int
	WriteBufferSize int
	ReadTimeout time.Duration
	WriteTimeout time.Duration
	MaxConnsPerIP int
	MaxRequestsPerConn int
	MaxKeepaliveDuration time.Duration
	TCPKeepalive bool
	TCPKeepalivePeriod time.Duration
	MaxRequestBodySize int
	DisableHeaderNamesNormalizing bool
	NoDefaultServerHeader bool
	NoDefaultContentType bool
}

func (cfg *serverConfig) makeServer() *fasthttp.Server {
	return nil
}

func (c *fastContext) InspectServerBlocks(path string, sblocks []caddyfile.ServerBlock) ([]caddyfile.ServerBlock, error) {
	return sblocks, nil
}

func (c *fastContext) MakeServers() ([]caddy.Server, error) {
	return []caddy.Server{NewFastServer(c.cfg),}, nil
}
