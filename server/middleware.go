package server

import (
	"github.com/valyala/fasthttp"
)

type Middleware func(handler fasthttp.RequestHandler) fasthttp.RequestHandler

func compileMiddleware(mList []Middleware, final fasthttp.RequestHandler) fasthttp.RequestHandler {
	stack := final
	for _, m := range mList {
		stack = m(stack)
	}
	return stack
}

func emptyMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return next
}

func notFoundHandler(ctx *fasthttp.RequestCtx) {
	ctx.NotFound()
}