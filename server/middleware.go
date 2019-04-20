package server

import (
	"github.com/valyala/fasthttp"
	"github.com/google/uuid"
)

type Middleware func(handler fasthttp.RequestHandler) fasthttp.RequestHandler

func compileMiddlewareEndWithNotFound(mList []Middleware, cfg NotFoundConfig) fasthttp.RequestHandler {
	final := notFound
	if cfg.StatusCode != 0 {
		final= newNotFoundHandler(cfg)
	}
	return compileMiddleware(mList, final)
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

func newNotFoundHandler(cfg NotFoundConfig) fasthttp.RequestHandler {
	if cfg.Body != "" {
		return func(ctx *fasthttp.RequestCtx) {
			ctx.Response.Reset()
			ctx.SetStatusCode(cfg.StatusCode)
			ctx.SetContentType(cfg.ContentType)
			ctx.SetBodyString(cfg.Body)
		}
	}
	// not_found directive's setup function must ensure the cfg.File is valid and accessible
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SendFile(cfg.File)
	}
}

func NewGzipMiddleware(cfg GzipConfig) Middleware {
	if !cfg.Open {
		return emptyMiddleware
	}
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return fasthttp.CompressHandlerLevel(next, cfg.Level)
	}
}

func NewRequestIDMiddleware(headerName string) Middleware {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			ctx.Request.Header.Set(headerName, uuid.New().String())
			ctx.SetUserValue(RequestIDHeaderName, headerName)
			next(ctx)
		}
	}
}