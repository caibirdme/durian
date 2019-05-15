package replace

import (
	"bytes"
	"errors"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type VariablePlaceholder struct {
	templates sync.Map
}

// NewVariablePlaceholder ...
func NewVariablePlaceholder() *VariablePlaceholder {
	return &VariablePlaceholder{templates: sync.Map{}}
}

func (vp *VariablePlaceholder) GetTmpl(tmplName string) (*fasttemplate.Template, bool) {
	v, ok := vp.templates.Load(tmplName)
	if !ok {
		return nil, false
	}
	return v.(*fasttemplate.Template), true
}

type TagFunc = fasttemplate.TagFunc

func (vp *VariablePlaceholder) ExecuteFuncString(tmplName string, fn TagFunc) (string, error) {
	tmpl, ok := vp.templates.Load(tmplName)
	if !ok {
		return "", ErrTmplNotExist
	}
	return tmpl.(*fasttemplate.Template).ExecuteFuncString(fn), nil
}

func (vp *VariablePlaceholder) ExecuteString(tmplName string, ctx *fasthttp.RequestCtx) (string, error) {
	return vp.ExecuteFuncString(tmplName, func(w io.Writer, tag string) (int, error) {
		return ReplaceVariable(ctx, w, tag)
	})
}

func (vp *VariablePlaceholder) SetTmpl(tmplName string) {
	vp.templates.Store(tmplName, fasttemplate.New(tmplName, "{", "}"))
}

// ReplaceVariable should be used in ExecuteFuncString
func ReplaceVariable(ctx *fasthttp.RequestCtx, w io.Writer, tag string) (int, error) {
	if f, ok := placeHolders[tag]; ok {
		return f(ctx, w)
	}
	switch tag[0] {
	case '>':
		return requestHeaderPlacer(ctx, w, tag[1:])
	case '<':
		return responseHeaderPlacer(ctx, w, tag[1:])
	case '?':
		return queryKeyPlacer(ctx, w, tag[1:])
	case '~':
		return cookiePlacer(ctx, w, tag[1:])
	}
	return 0, ErrNotBuiltin
}

type ReplaceFunc func(ctx *fasthttp.RequestCtx, w io.Writer) (int, error)

var placeHolders = map[string]ReplaceFunc{
	"host":         hostPlacer,
	"hostonly":     hostonlyPlacer,
	"hostname":     hostnamePlacer,
	"method":       methodPlacer,
	"path":         pathPlacer,
	"proto":        protoPlacer,
	"query":        queryPlacer,
	"remote":       remotePlacer,
	"port":         portPlacer,
	"schema":       schemaPlacer,
	"uri":          uriPlacer,
	"when_iso":     whenISOPlacer,
	"when_unix":    whenUnixPlacer,
	"when_unix_ms": whenUnixMsPlacer,
	"fragment":     fragmentPlacer,
	"latency":      latencyPlacer,
	"latency_ms":   latencyMsPlacer,
	"status":       statusPlacer,
}

func statusPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write([]byte(strconv.Itoa(ctx.Response.StatusCode())))
}

func latencyMsPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write([]byte(strconv.FormatInt(time.Since(ctx.Time()).Nanoseconds()/1e6, 10)))
}

func latencyPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write([]byte(time.Since(ctx.Time()).String()))
}

func cookiePlacer(ctx *fasthttp.RequestCtx, w io.Writer, key string) (int, error) {
	return w.Write(ctx.Request.Header.Cookie(key))
}

func queryKeyPlacer(ctx *fasthttp.RequestCtx, w io.Writer, key string) (int, error) {
	return w.Write(ctx.QueryArgs().Peek(key))
}

func responseHeaderPlacer(ctx *fasthttp.RequestCtx, w io.Writer, header string) (int, error) {
	return w.Write(ctx.Response.Header.Peek(header))
}

func requestHeaderPlacer(ctx *fasthttp.RequestCtx, w io.Writer, header string) (int, error) {
	return w.Write(ctx.Request.Header.Peek(header))
}

func fragmentPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write(ctx.URI().Hash())
}

func whenUnixMsPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write([]byte(strconv.FormatInt(ctx.Time().UnixNano()/1e6, 10)))
}

func whenUnixPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write([]byte(strconv.FormatInt(ctx.Time().Unix(), 10)))
}

func whenISOPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write([]byte(ctx.Time().Format(time.RFC3339)))
}

func uriPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write(ctx.URI().RequestURI())
}

func schemaPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write(ctx.URI().Scheme())
}

func portPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write([]byte(strconv.Itoa(ctx.RemoteAddr().(*net.TCPAddr).Port)))
}

func remotePlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write([]byte(ctx.RemoteIP().String()))
}

func queryPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write(ctx.URI().QueryString())
}

func protoPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	if ctx.Request.Header.IsHTTP11() {
		return w.Write(strHTTP11)
	}
	return w.Write(strHTTP10)
}

func hostPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write(ctx.Host())
}

var (
	_hostname []byte
	strHTTP11 = []byte("HTTP/1.1")
	strHTTP10 = []byte("HTTP/1.0")
	// ErrNotBuiltin indicate the value of placeholder isn't builtin, users need to implement themselves
	ErrNotBuiltin = errors.New("not built-in variable")
	// ErrTmplNotExist ...
	ErrTmplNotExist = errors.New("tmpl not exist")
)

func init() {
	h, err := os.Hostname()
	if err == nil {
		_hostname = []byte(h)
	}
}

func hostonlyPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	h := ctx.Host()
	idx := bytes.IndexByte(h, ':')
	if idx == -1 {
		return w.Write(h)
	}
	return w.Write(h[:idx])
}

func hostnamePlacer(_ *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write(_hostname)
}

func methodPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write(ctx.Method())
}

func pathPlacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	return w.Write(ctx.Path())
}
