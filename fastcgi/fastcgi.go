package fastcgi

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/caibirdme/durian/log"
	"github.com/caibirdme/durian/replace"
	super "github.com/caibirdme/durian/server"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Handler struct {
	Debug       bool
	KeepConn    bool
	ReadTimeout time.Duration
	SendTimeout time.Duration
	rule        *Rule
	Next        fasthttp.RequestHandler
	UpstreamGetter
}

func NewHandler(rule *Rule, cfg *Config, next fasthttp.RequestHandler) (*Handler, error) {
	upg, err := NewRoundRobinUpstream(cfg.Upstream)
	if err != nil {
		return nil, err
	}
	h := &Handler{
		UpstreamGetter: upg,
		rule:           rule,
		Next:           next,
		ReadTimeout:    cfg.ReadTimeout,
		SendTimeout:    cfg.SendTimeout,
		KeepConn:       cfg.KeepConn,
		Debug:          cfg.Debug,
	}
	return h, nil
}

func (h *Handler) Serve(reqCtx *fasthttp.RequestCtx) {
	if !h.rule.location.Match(reqCtx.Path()) {
		h.Next(reqCtx)
		return
	}
	pathInfo, err := h.rule.newPathInfo(reqCtx.Request.URI().Path())
	if err != nil {
		reqCtx.Error("[fcgi] path split error", fasthttp.StatusInternalServerError)
		if h.Debug {
			log.GetLogger().Error("[fcgi] path split error", zap.Error(err), zap.ByteString("path", reqCtx.Request.URI().Path()))
		}
		return
	}
	reqCtx.SetUserValue(pathInfoKey, pathInfo)
	env, err := h.rule.buildEnv(reqCtx)
	if err != nil {
		reqCtx.Error("[fcgi] internal error", fasthttp.StatusInternalServerError)
		if h.Debug {
			log.GetLogger().Error("[fcgi] buildEnv error",
				zap.Error(err),
				zap.String("scriptName", pathInfo.ScriptName),
				zap.String("pathInfo", pathInfo.PathInfo),
				zap.ByteString("path", reqCtx.Request.URI().Path()),
			)
		}
		return
	}
	network, addr := h.GetAddress()
	fcgi, err := h.getFCGIClient(reqCtx, network, addr)
	if err != nil {
		reqCtx.Error("[fcgi] fail to connect backend", fasthttp.StatusBadGateway)
		if h.Debug {
			log.GetLogger().Error("[fcgi] fail to connect backend",
				zap.Error(err),
				zap.String("network", network),
				zap.String("addr", addr),
			)
		}
		return
	}
	var resp *http.Response
	switch string(reqCtx.Method()) {
	case strGet:
		resp, err = fcgi.Get(env, bytes.NewReader(reqCtx.Request.Body()), int64(reqCtx.Request.Header.ContentLength()))
	case strHead:
		resp, err = fcgi.Head(env)
	case strOptions:
		resp, err = fcgi.Options(env)
	default:
		resp, err = fcgi.Post(
			env,
			string(reqCtx.Method()),
			string(reqCtx.Request.Header.ContentType()),
			bytes.NewReader(reqCtx.Request.Body()),
			int64(reqCtx.Request.Header.ContentLength()),
		)
	}
	if err != nil {
		reqCtx.Error("[fcgi] request backend error", fasthttp.StatusBadGateway)
		if h.Debug {
			log.GetLogger().Error("[fcgi] fail to connect backend", zap.Error(err))
		}
		return
	}
	for k, v := range resp.Header {
		reqCtx.Response.Header.Set(k, v[0])
	}
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	reqCtx.Write(body)
	reqCtx.SetStatusCode(resp.StatusCode)
}

const (
	strGet      = "GET"
	strHead     = "HEAD"
	strOptions  = "OPTIONS"
	pathInfoKey = "_path_info"
)

func (h *Handler) getFCGIClient(reqCtx *fasthttp.RequestCtx, network, addr string) (*FCGIClient, error) {
	ctx := super.GetStdCtx(reqCtx)
	fcgi, err := DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	err = fcgi.SetReadTimeout(h.ReadTimeout)
	if err != nil {
		return nil, err
	}
	err = fcgi.SetSendTimeout(h.SendTimeout)
	if err != nil {
		return nil, err
	}
	return fcgi, nil
}

type Rule struct {
	location       super.LocationMatcher
	Root           string
	Index          string
	SplitPathInfo  *regexp.Regexp
	CatchStderr    string
	Params         map[string]string
	ServerSoftware string
	ServerName     string
	templates      *replace.VariablePlaceholder
}

type pathInfo struct {
	ScriptName string
	PathInfo   string
}

var (
	errSplitFail  = errors.New("fail to split path")
	_headerPrefix = []byte("HTTP_")
)

func addExt(path string, ext string) string {
	if strings.HasSuffix(path, ext) {
		return path
	}
	if path[len(path)-1] == '/' {
		return path + ext
	}
	return path + "/" + ext
}

func (r *Rule) newPathInfo(path []byte) (pathInfo, error) {
	if r.SplitPathInfo == nil {
		return pathInfo{
			ScriptName: addExt(string(path), r.Index),
		}, nil
	}
	matched := r.SplitPathInfo.FindSubmatch(path)
	if len(matched) < 2 {
		return pathInfo{}, errSplitFail
	}
	pInfo := pathInfo{
		ScriptName: addExt(string(matched[1]), r.Index),
	}
	if len(matched) >= 3 {
		pInfo.PathInfo = string(matched[2])
	}
	return pInfo, nil
}

func (r *Rule) buildEnv(ctx *fasthttp.RequestCtx) (map[string]string, error) {
	env := make(map[string]string)
	// 详见nginx文档： https://www.nginx.com/resources/wiki/start/topics/examples/phpfcgi/
	env["AUTH_TYPE"] = ""
	env["QUERY_STRING"] = string(ctx.URI().QueryString())
	env["REQUEST_METHOD"] = string(ctx.Method())
	env["CONTENT_TYPE"] = string(ctx.Request.Header.ContentType())
	env["CONTENT_LENGTH"] = strconv.Itoa(ctx.Request.Header.ContentLength())
	env["REQUEST_URI"] = string(ctx.Request.URI().RequestURI())
	env["DOCUMENT_URI"] = string(ctx.Request.URI().Path())
	env["DOCUMENT_ROOT"] = r.Root
	serverProtocol := "HTTP/1.1"
	if !ctx.Request.Header.IsHTTP11() {
		serverProtocol = "HTTP/1.0"
	}
	env["SERVER_PROTOCOL"] = serverProtocol
	env["GATEWAY_INTERFACE"] = "CGI/1.1"
	env["SERVER_SOFTWARE"] = r.ServerSoftware
	ip, port, err := getAddr(ctx.RemoteAddr())
	if err != nil {
		return nil, err
	}
	env["REMOTE_ADDR"] = ip
	env["REMOTE_PORT"] = port
	ip, port, err = getAddr(ctx.LocalAddr())
	if err != nil {
		return nil, err
	}
	env["SERVER_ADDR"] = ip
	env["SERVER_PORT"] = port
	env["SERVER_NAME"] = r.ServerName

	// set custom env
	for k, v := range r.Params {
		realV, err := r.DoReplace(ctx, v)
		if err != nil {
			// todo: log
		}
		env[k] = realV
	}
	if path_info := env["PATH_INFO"]; path_info == "" {
		delete(env, "PATH_TRANSLATED")
		delete(env, "PATH_INFO")
	}

	// set header
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		var sb strings.Builder
		sb.Write(_headerPrefix)
		sb.Write(convertHeader2EnvParam(k))
		env[sb.String()] = string(v)
	})

	return env, nil
}

func (r *Rule) DoReplace(ctx *fasthttp.RequestCtx, tmplName string) (string, error) {
	return r.templates.ExecuteFuncString(tmplName, func(w io.Writer, tag string) (int, error) {
		written, err := r.ReplaceVariable(ctx, w, tag)
		if err == nil {
			return written, nil
		} else if err != _errNotBuiltin {
			return written, err
		}
		return replace.ReplaceVariable(ctx, w, tag)
	})
}

var (
	_errNotBuiltin = errors.New("not fastcgi built-in variables")
)

func (r *Rule) ReplaceVariable(ctx *fasthttp.RequestCtx, w io.Writer, tag string) (int, error) {
	if f, ok := placeHolders[tag]; ok {
		return f(ctx, w)
	}
	switch tag {
	case "root":
		return w.Write([]byte(r.Root))
	default:
		return 0, _errNotBuiltin
	}
}

func (r *Rule) includeScriptParam() {
	if _, ok := r.Params["SCRIPT_NAME"]; !ok {
		r.Params["SCRIPT_NAME"] = "{fastcgi_script_name}"
		r.templates.SetTmpl("{fastcgi_script_name}")
	}
	if _, ok := r.Params["SCRIPT_FILENAME"]; !ok {
		r.Params["SCRIPT_FILENAME"] = "{root}{fastcgi_script_name}"
		r.templates.SetTmpl("{root}{fastcgi_script_name}")
	}
	if _, ok := r.Params["PATH_INFO"]; !ok {
		r.Params["PATH_INFO"] = "{fastcgi_path_info}"
		r.templates.SetTmpl("{fastcgi_path_info}")
	}
	if _, ok := r.Params["PATH_TRANSLATED"]; !ok {
		r.Params["PATH_TRANSLATED"] = "{root}{fastcgi_path_info}"
		r.templates.SetTmpl("{root}{fastcgi_path_info}")
	}
}

var placeHolders = map[string]replace.ReplaceFunc{
	"fastcgi_script_name": scriptNameReplacer,
	"fastcgi_path_info":   pathInfoReplacer,
}

func scriptNameReplacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	pInfo := ctx.UserValue(pathInfoKey).(pathInfo)
	return w.Write([]byte(pInfo.ScriptName))
}

func pathInfoReplacer(ctx *fasthttp.RequestCtx, w io.Writer) (int, error) {
	pInfo := ctx.UserValue(pathInfoKey).(pathInfo)
	return w.Write([]byte(pInfo.PathInfo))
}

func convertHeader2EnvParam(k []byte) []byte {
	return bytes.Replace(bytes.ToUpper(k), []byte{'-'}, []byte{'_'}, -1)
}

func getAddr(addr net.Addr) (string, string, error) {
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		return tcpAddr.IP.String(), strconv.Itoa(tcpAddr.Port), nil
	}
	return "", "", fmt.Errorf("[fastcgi] fail to getAddr: %s", addr.String())
}

type UpstreamGetter interface {
	GetAddress() (network string, address string)
}

type roundRobinUpstream struct {
	backends []super.Backend
	pos      int64
}

func (rr *roundRobinUpstream) GetAddress() (network string, address string) {
	n := len(rr.backends)
	new_pos := atomic.AddInt64(&rr.pos, 1)
	idx := int(new_pos % int64(n))
	return rr.backends[idx].Network, rr.backends[idx].Addr
}

func NewRoundRobinUpstream(upstream super.Upstream) (UpstreamGetter, error) {
	return &roundRobinUpstream{
		backends: upstream.Backends,
	}, nil
}
