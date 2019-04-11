package reverse_proxy

import (
	"time"

	"github.com/valyala/fasthttp"
)

func ProxyHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(reqCtx *fasthttp.RequestCtx) {
	}
}

type Proxy struct {
	next    fasthttp.RequestHandler
	lb      *fasthttp.LBClient
	check   URLMatchChecker
	timeout time.Duration
}

func NewProxy(next fasthttp.RequestHandler, servers []string, pattern string, timeout time.Duration) (*Proxy, error) {
	checker, err := NewRegexpMatcher(pattern)
	if nil != err {
		return nil, err
	}
	var lbc fasthttp.LBClient
	for _, addr := range servers {
		c := &fasthttp.HostClient{
			Addr: addr,
		}
		lbc.Clients = append(lbc.Clients, c)
	}
	return &Proxy{
		lb:      &lbc,
		check:   checker,
		timeout: timeout,
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
		err := p.lb.DoTimeout(req, resp, p.timeout)
		if err == fasthttp.ErrTimeout {
			reqCtx.TimeoutError(err.Error())
		} else {
			reqCtx.Error(err.Error(), fasthttp.StatusServiceUnavailable)
		}
	}
}
