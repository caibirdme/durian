package caddy_fasthttp

import (
	"github.com/valyala/fasthttp"
)

type Middleware func(handler fasthttp.RequestHandler) fasthttp.RequestHandler

func compileMiddleware(mList []Middleware) fasthttp.RequestHandler {
	stack := notFound
	for _,m := range mList {
		stack = m(stack)
	}
	return stack
}

func notFound(reqCtx *fasthttp.RequestCtx) {
	reqCtx.NotFound()
}