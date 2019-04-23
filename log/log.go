package log

import (
	"fmt"
	super "github.com/caibirdme/durian/server"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultAccessLogName = "access.log"
	defaultErrLogName    = "error.log"
)

func NewLogger(cfg LogConfig) (EntityWriter, func() error, error) {
	err := confirmPath(&cfg)
	if err != nil {
		return nil, nil, err
	}
	lg, err := newZapLogger(cfg)
	if err != nil {
		return nil, nil, err
	}
	fwriter, err := newFormatWriter(lg, cfg.Format)
	if err != nil {
		return nil, nil, err
	}
	return fwriter, lg.Sync, nil
}

func confirmPath(cfg *LogConfig) error {
	if filepath.Ext(cfg.AccessPath) == "" {
		cfg.AccessPath = filepath.Join(cfg.AccessPath, defaultAccessLogName)
	}
	dir := filepath.Dir(cfg.AccessPath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	if filepath.Ext(cfg.ErrPath) == "" {
		cfg.ErrPath = filepath.Join(cfg.ErrPath, defaultErrLogName)
	}
	dir = filepath.Dir(cfg.ErrPath)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return nil
}

func newZapLogger(cfg LogConfig) (*zap.Logger, error) {
	zapCfg := zap.Config{
		Level:       zap.NewAtomicLevel(),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    newProductionEncoderConfig(),
		OutputPaths:      []string{cfg.AccessPath},
		ErrorOutputPaths: []string{cfg.ErrPath},
	}
	return zapCfg.Build()
}

func newProductionEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

type EntityWriter interface {
	Write(ctx *fasthttp.RequestCtx)
}

type formatWriter struct {
	l       *zap.Logger
	writers []partialWriter
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]zapcore.Field, 0, 6)
	},
}

func (f *formatWriter) Write(ctx *fasthttp.RequestCtx) {
	fields := bufPool.Get().([]zapcore.Field)
	fields = fields[:0]
	for _, h := range f.writers {
		fields = append(fields, h(ctx))
	}
	f.l.Info("", fields...)
	bufPool.Put(fields)
}

func newFormatWriter(l *zap.Logger, format []string) (*formatWriter, error) {
	writers := make([]partialWriter, 0, len(format))
	for _, val := range format {
		h, ok := writerDict[val]
		if !ok {
			return nil, fmt.Errorf("[log] %s not supported", val)
		}
		writers = append(writers, h)
	}
	return &formatWriter{
		l:       l,
		writers: writers,
	}, nil
}

//  caller's responsibility to ensure entity isn't nil
type partialWriter func(ctx *fasthttp.RequestCtx) zapcore.Field

const (
	entryStartTime             = "start_time"
	entryKeyBytesSent          = "bytes_sent"
	entryKeyBodyBytesSent      = "body_bytes_sent"
	entryKeyConnectionRequests = "connection_requests"
	entryKeyProcessTime        = "process_time"
	entryKeyRequestLength      = "request_length"
	entryKeyStatusCode         = "status"
	entryKeyUA                 = "user_agent"
	entryKeyRemoteAddr         = "remote_addr"
	entryKeyRequestURI         = "request_uri"
	entryKeyQueryString        = "query_string"
	entryKeyRequestBody        = "request_body"
	entryKeyRequestHeader      = "request_header"
	entryKeyMethod             = "method"
	entryKeyResponseBody       = "response_body"
	entryKeyResponseHeader     = "response_header"
	entryReferer               = "referer"
	entryRequestID             = "request_id"
	entryKeyHost               = "host"
)

var (
	writerDict = map[string]partialWriter{
		entryKeyHost:               hostWriter,
		entryRequestID:             requestIDWriter,
		entryStartTime:             startTimeWriter,
		entryReferer:               refererWriter,
		entryKeyBytesSent:          bytesSentWriter,
		entryKeyBodyBytesSent:      bodyBytesSentWriter,
		entryKeyConnectionRequests: connectionRequestsWriter,
		entryKeyProcessTime:        processTimeWriter,
		entryKeyRequestLength:      requestLengthWriter,
		entryKeyStatusCode:         statusCodeWriter,
		entryKeyUA:                 userAgentWriter,
		entryKeyRemoteAddr:         remoteAddrWriter,
		entryKeyRequestURI:         requestURIWriter,
		entryKeyQueryString:        queryStringWriter,
		entryKeyRequestBody:        requestBodyWriter,
		entryKeyRequestHeader:      requestHeaderWriter,
		entryKeyMethod:             methodWriter,
		entryKeyResponseBody:       responseBodyWriter,
		entryKeyResponseHeader:     responseHeaderWriter,
	}
)

func hostWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.ByteString(entryKeyHost, ctx.Host())
}

func requestIDWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	headerKey := ctx.UserValue(super.RequestIDHeaderName)
	var reqID []byte
	if headerKey != nil {
		reqID = ctx.Request.Header.Peek(headerKey.(string))
	}
	if len(reqID) == 0 {
		return zap.String(entryRequestID, "-")
	} else {
		return zap.ByteString(entryRequestID, reqID)
	}
}

func refererWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.ByteString(entryReferer, ctx.Referer())
}

func startTimeWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.Time(entryStartTime, ctx.Time())
}

func responseHeaderWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.ByteString(entryKeyResponseHeader, ctx.Response.Header.Header())
}

func responseBodyWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.ByteString(entryKeyResponseBody, ctx.Response.Body())
}

func methodWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.ByteString(entryKeyMethod, ctx.Method())
}

func requestHeaderWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.ByteString(entryKeyRequestHeader, ctx.Request.Header.Header())
}

func requestBodyWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.ByteString(entryKeyRequestBody, ctx.Request.Body())
}

func queryStringWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.ByteString(entryKeyQueryString, ctx.Request.URI().QueryString())
}

func requestURIWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.ByteString(entryKeyRequestURI, ctx.Request.URI().Path())
}

func remoteAddrWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.String(entryKeyRemoteAddr, ctx.RemoteAddr().String())
}

func userAgentWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.ByteString(entryKeyUA, ctx.Request.Header.UserAgent())
}

func bytesSentWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	count := len(ctx.Response.Header.Header()) + len(ctx.Response.Body())
	return zap.Int(entryKeyBytesSent, count)
}

func bodyBytesSentWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.Int(entryKeyBodyBytesSent, len(ctx.Response.Body()))
}

func connectionRequestsWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.Uint64(entryKeyConnectionRequests, ctx.ConnRequestNum())
}

func processTimeWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.Duration(entryKeyProcessTime, time.Since(ctx.Time()))
}

func requestLengthWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	count := len(ctx.Request.Header.Header()) + len(ctx.Request.Body())
	return zap.Int(entryKeyRequestLength, count)
}

func statusCodeWriter(ctx *fasthttp.RequestCtx) zapcore.Field {
	return zap.Int(entryKeyStatusCode, ctx.Response.StatusCode())
}
