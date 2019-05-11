package server

import (
	"context"
	"github.com/valyala/fasthttp"
)

type KVTuple struct {
	K, V string
}

const (
	// StandardContextKey use this key to access context.Context from ctx.UserValues
	standardContextKey = "_ctx"
	DurianName         = "durian"
)

type StorageKey int

const (
	UpstreamKey StorageKey = iota
	ServerNameKey
	DocRootKey
)

func GetStdCtx(reqCtx *fasthttp.RequestCtx) context.Context {
	ctx, ok := reqCtx.UserValue(standardContextKey).(context.Context)
	if ok {
		return ctx
	}
	return context.TODO()
}

type Upstream struct {
	Name     string
	Backends []Backend
}

type Backend struct {
	Network string
	Addr    string
	Weight  int
	Backup  bool
}
