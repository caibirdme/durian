package server

import (
	"github.com/valyala/fasthttp"
)

type Middleware func(handler fasthttp.RequestHandler) fasthttp.RequestHandler

func compileMiddlewareEndWithNotFound(mList []Middleware) fasthttp.RequestHandler {
	return compileMiddleware(mList, notFound)
}

func compileMiddleware(mList []Middleware, final fasthttp.RequestHandler) fasthttp.RequestHandler {
	stack := final
	for _, m := range mList {
		stack = m(stack)
	}
	return stack
}

func notFound(reqCtx *fasthttp.RequestCtx) {
	reqCtx.NotFound()
}

func emptyMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return next
}

func NewGzipMiddleware(cfg GzipConfig) Middleware {
	if !cfg.Open {
		return emptyMiddleware
	}
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return fasthttp.CompressHandlerLevel(next, cfg.Level)
	}
}
