package reverse_proxy

import (
	"strings"
	"time"

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
	headerUpstream   []KVTuple
	headerDownstream []KVTuple
}

func NewProxy(cfg ProxyConfig) (*Proxy, error) {
	checker, err := NewRegexpMatcher(cfg.Pattern)
	if nil != err {
		return nil, err
	}
	addr := strings.Join(cfg.AddressList, ",")
	return &Proxy{
		client: &fasthttp.HostClient{
			Addr: addr,
		},
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
		req := &reqCtx.Request
		resp := &reqCtx.Response
		for _, tuple := range p.headerUpstream {
			req.Header.Set(tuple.K, tuple.V)
		}
		err := p.client.DoTimeout(req, resp, p.timeout)
		if err != nil {
			if err == fasthttp.ErrTimeout {
				reqCtx.TimeoutError(err.Error())
			} else {
				reqCtx.Error(err.Error(), fasthttp.StatusServiceUnavailable)
			}
		}
		for _, tuple := range p.headerDownstream {
			resp.Header.Set(tuple.K, tuple.V)
		}
	}
}
