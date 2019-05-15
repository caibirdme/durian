package header

import (
	"github.com/caibirdme/durian/replace"
	super "github.com/caibirdme/durian/server"
	"github.com/valyala/fasthttp"
)

type HeaderSetter interface {
	Set(ctx *fasthttp.RequestCtx) error
}

type headerSetter struct {
	tuples    []super.KVTuple
	templates *replace.VariablePlaceholder
}

func (h *headerSetter) Set(ctx *fasthttp.RequestCtx) error {
	for _, pair := range h.tuples {
		val, err := h.templates.ExecuteString(pair.V, ctx)
		if err != nil {
			return err
		}
		ctx.Request.Header.Set(pair.K, val)
	}
	return nil
}

func NewHeaderSetter(tuples []super.KVTuple) HeaderSetter {
	h := headerSetter{
		tuples:    tuples,
		templates: replace.NewVariablePlaceholder(),
	}
	for _, pair := range tuples {
		h.templates.SetTmpl(pair.V)
	}
	return &h
}
