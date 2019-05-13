package fastcgi

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/caibirdme/durian/log"
	super "github.com/caibirdme/durian/server"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
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
	if !h.rule.Match(reqCtx) {
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
	env, err := h.rule.buildEnv(reqCtx, &pathInfo)
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
	fmt.Println("----", env)
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
	strGet     = "GET"
	strHead    = "HEAD"
	strOptions = "OPTIONS"
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
	Pattern        *regexp.Regexp
	Root           string
	Index          string
	SplitPathInfo  *regexp.Regexp
	ScriptFileName string
	CatchStderr    string
	Params         map[string]string
	ServerSoftware string
	ServerName     string
	templates      sync.Map
}

func (r *Rule) Match(ctx *fasthttp.RequestCtx) bool {
	return r.Pattern.Match(ctx.RequestURI())
}

type pathInfo struct {
	ScriptName string
	PathInfo   string
}

var (
	errSplitFail = errors.New("fail to split path")
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

func (r *Rule) getScriptFileName(scriptName string) string {
	if r.ScriptFileName == "" {
		var pre string
		if r.Root[len(r.Root)-1] != '/' {
			pre = r.Root + "/"
		} else {
			pre = r.Root
		}
		return pre + scriptName
	}
	scriptFileName := strings.Replace(r.ScriptFileName, "$document_root", r.Root, 1)
	return strings.Replace(scriptFileName, "$fastcgi_script_name", scriptName, 1)
}

func (r *Rule) buildEnv(ctx *fasthttp.RequestCtx, pathInfo *pathInfo) (map[string]string, error) {
	env := make(map[string]string)
	// 详见nginx文档： https://www.nginx.com/resources/wiki/start/topics/examples/phpfcgi/
	env["AUTH_TYPE"] = ""
	env["QUERY_STRING"] = string(ctx.URI().QueryString())
	env["REQUEST_METHOD"] = string(ctx.Method())
	env["CONTENT_TYPE"] = string(ctx.Request.Header.ContentType())
	env["CONTENT_LENGTH"] = strconv.Itoa(ctx.Request.Header.ContentLength())
	env["SCRIPT_FILENAME"] = r.getScriptFileName(pathInfo.ScriptName)
	env["SCRIPT_NAME"] = pathInfo.ScriptName
	env["PATH_INFO"] = pathInfo.PathInfo
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
	return env, nil
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
