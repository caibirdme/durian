package log

import (
	"errors"
	"fmt"
	super "github.com/caibirdme/durian/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
	"time"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveLog, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setupAccess,
	})
}

func setupAccess(c *caddy.Controller) error {
	fmt.Println("setup log")
	cfg, err := parseConfig(c)
	if err != nil {
		return err
	}
	logWriter, sync, err := NewLogger(*cfg)
	if nil != err {
		return fmt.Errorf("[log] init log error: %s", err)
	}
	if sync != nil {
		c.OnShutdown(sync)
	}
	super.GetConfig(c).AddNamedMiddleware(super.LogMiddlewareName, func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			next(ctx)
			logWriter.Write(ctx)
		}
	})
	return nil
}

func parseConfig(c *caddy.Controller) (*LogConfig, error) {
	c.Next()
	cfg := LogConfig{}
	block := false
	for c.NextBlock() {
		block = true
		kind := c.Val()
		switch strings.ToLower(kind) {
		case "access_path":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			cfg.AccessPath = c.Val()
		case "err_path":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			cfg.ErrPath = c.Val()
		case "format":
			format, err := parseFormat(c)
			if err != nil {
				return nil, err
			}
			cfg.Format = format
		case "buffer":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			size, err := parseSize(c.Val())
			if nil != err {
				return nil, c.Err(err.Error())
			}
			cfg.Buffer = size
		case "flush":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			d, err := time.ParseDuration(c.Val())
			if nil != err {
				return nil, c.Err(err.Error())
			}
			cfg.Flush = d
		}
	}
	if !block {
		return nil, c.Err("log related config should be written in a block")
	}
	return &cfg, nil
}

func parseSize(s string) (int, error) {
	s = strings.Trim(s, " ")
	size, err := strconv.Atoi(s[:len(s)-1])
	if nil != err {
		return 0, err
	}
	unit := s[len(s)-1]
	switch unit {
	case 'k', 'K':
		return size * 1024, nil
	case 'm', 'M':
		return size * 1024 * 1024, nil
	case 'g', 'G':
		return size * 1024 * 1024 * 1024, nil
	default:
		return 0, errors.New("invalid unit")
	}
}

var (
	defaultFormat = []string{
		entryKeyRemoteAddr,
		entryKeyHost,
		entryKeyMethod,
		entryKeyRequestURI,
		entryKeyStatusCode,
		entryStartTime,
		entryKeyProcessTime,
		entryKeyBytesSent,
		entryReferer,
		entryKeyUA,
	}
)

func parseFormat(c *caddy.Controller) ([]string, error) {
	if !c.NextArg() {
		return defaultFormat, nil
	}
	if c.Val() != "{" {
		return nil, c.ArgErr()
	}
	arr := make([]string, 0, 6)
	hash := make(map[string]struct{})
	foundRightParen := false
	for c.Next() {
		if c.Val() == "}" {
			foundRightParen = true
			break
		}
		if _, ok := hash[c.Val()]; !ok {
			arr = append(arr, c.Val())
			hash[c.Val()] = struct{}{}
		}
	}
	if !foundRightParen {
		return nil, c.Err("not found right paren")
	}
	return arr, nil
}

type LogConfig struct {
	AccessPath string
	ErrPath    string
	Format     []string
	Buffer     int
	Flush      time.Duration
}
