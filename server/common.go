package server

import (
	"bytes"
	"context"
	"errors"
	"github.com/valyala/fasthttp"
	"regexp"
)

type KVTuple struct {
	K, V string
}

const (
	// StandardContextKey use this key to access context.Context from ctx.UserValues
	standardContextKey = "_ctx"
	DurianName         = "durian"
	DurianVersion      = "0.0.1"
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

type location struct {
	pattern *regexp.Regexp
	prefix  []byte
}

func (lo *location) Match(uri []byte) bool {
	if lo.pattern != nil {
		return lo.pattern.Match(uri)
	}
	return bytes.HasPrefix(uri, lo.prefix)
}

type LocationMatcher interface {
	Match(uri []byte) bool
}

func NewLocationMatcher(firstLine []string) (LocationMatcher, error) {
	if len(firstLine) == 0 {
		return nil, errors.New("nil firstLine")
	}
	if len(firstLine) == 1 {
		return &location{prefix: []byte(firstLine[0])}, nil
	}
	re, err := regexp.Compile(firstLine[1])
	if err != nil {
		return nil, err
	}
	return &location{pattern: re}, nil
}

type combineMather struct {
	should  LocationMatcher
	exclude LocationMatcher
}

func (c *combineMather) Match(uri []byte) bool {
	if c.should.Match(uri) && !c.exclude.Match(uri) {
		return true
	}
	return false
}

func NewCombineMatcher(shouldMatch LocationMatcher, exclude LocationMatcher) LocationMatcher {
	return &combineMather{should: shouldMatch, exclude: exclude}
}
