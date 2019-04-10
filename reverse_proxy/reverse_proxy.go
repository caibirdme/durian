package reverse_proxy

import (
	"github.com/valyala/fasthttp"
)

func ProxyHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(reqCtx *fasthttp.RequestCtx) {
	}
}
