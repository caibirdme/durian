package reverse_proxy

import (
	"bytes"
	"regexp"
)

type URLMatchChecker interface {
	Match([]byte) bool
}

func NewRegexpMatcher(pattern string) (URLMatchChecker, error) {
	return regexp.Compile(pattern)
}

type PrefixChecker struct {
	prefix []byte
}

func (p *PrefixChecker) Match(urlPath []byte) bool {
	return bytes.HasPrefix(urlPath, p.prefix)
}

func NewPrefixChecker(prefix string) (URLMatchChecker,error) {
	return &PrefixChecker{prefix:[]byte(prefix),},nil
}