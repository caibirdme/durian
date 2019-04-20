package reverse_proxy

import (
	"strings"
	"time"

	super "github.com/caibirdme/durian/server"
	"github.com/valyala/fasthttp"
)

func ProxyHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(reqCtx *fasthttp.RequestCtx) {
	}
}

type Proxy struct {
	next             fasthttp.RequestHandler
	client           *fasthttp.HostClient
	check            URLMatchChecker
	timeout          time.Duration
	headerUpstream   []super.KVTuple
	headerDownstream []super.KVTuple
}

func NewProxy(cfg ProxyConfig) (*Proxy, error) {
	var checker URLMatchChecker
	var err error
	if cfg.Pattern != "" {
		checker, err = NewRegexpMatcher(cfg.Pattern)
	} else {
		checker, err = NewPrefixChecker(cfg.Path)
	}
	if nil != err {
		return nil, err
	}
	addr := strings.Join(cfg.AddressList, ",")
	client := &fasthttp.HostClient{
		Addr: addr,
	}
	if cfg.MaxConn > 0 {
		client.SetMaxConns(cfg.MaxConn)
	}
	return &Proxy{
		client:           client,
		check:            checker,
		timeout:          cfg.Timeout,
		headerUpstream:   cfg.UpstreamHeader,
		headerDownstream: cfg.DownstreamHeader,
	}, nil
}

func (p *Proxy) Handle(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(reqCtx *fasthttp.RequestCtx) {
		if !p.check.Match(reqCtx.Path()) {
			next(reqCtx)
			return
		}
		for _, tuple := range p.headerUpstream {
			reqCtx.Request.Header.Set(tuple.K, tuple.V)
		}

		err := p.client.DoTimeout(&reqCtx.Request, &reqCtx.Response, p.timeout)
		if err != nil {
			if err == fasthttp.ErrTimeout {
				reqCtx.TimeoutError(err.Error())
			} else {
				reqCtx.Error(err.Error(), fasthttp.StatusServiceUnavailable)
			}
		}
		for _, tuple := range p.headerDownstream {
			reqCtx.Response.Header.Set(tuple.K, tuple.V)
		}
	}
}
