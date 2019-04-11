package reverse_proxy

import (
	"regexp"
)

type URLMatchChecker interface {
	Match([]byte) bool
}

func NewRegexpMatcher(pattern string) (URLMatchChecker, error) {
	return regexp.Compile(pattern)
}
