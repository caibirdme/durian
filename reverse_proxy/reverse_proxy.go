package reverse_proxy

import (
	"log"
	"time"

	"github.com/valyala/fasthttp"
)

func ProxyHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(reqCtx *fasthttp.RequestCtx) {
	}
}

type Proxy struct {
	next             fasthttp.RequestHandler
	lb               *fasthttp.LBClient
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
	var lbc fasthttp.LBClient
	for _, addr := range cfg.AddressList {
		c := &fasthttp.HostClient{
			Addr: addr,
		}
		lbc.Clients = append(lbc.Clients, c)
	}
	return &Proxy{
		lb:               &lbc,
		check:            checker,
		timeout:          cfg.Timeout,
		headerUpstream:   cfg.UpstreamHeader,
		headerDownstream: cfg.DownstreamHeader,
	}, nil
}

func (p *Proxy) Handle(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(reqCtx *fasthttp.RequestCtx) {
		log.Printf("I'm here %s\n", reqCtx.Path())
		if !p.check.Match(reqCtx.Path()) {
			log.Printf("%s not match\n", reqCtx.Path())
			next(reqCtx)
			return
		}
		req := &reqCtx.Request
		resp := &reqCtx.Response
		for _, tuple := range p.headerUpstream {
			req.Header.Set(tuple.K, tuple.V)
		}
		err := p.lb.DoTimeout(req, resp, p.timeout)
		if err == fasthttp.ErrTimeout {
			reqCtx.TimeoutError(err.Error())
		} else {
			reqCtx.Error(err.Error(), fasthttp.StatusServiceUnavailable)
		}
		for _, tuple := range p.headerDownstream {
			resp.Header.Set(tuple.K, tuple.V)
		}
	}
}
