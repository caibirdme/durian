package fastcgi

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRule_newPathInfo(t *testing.T) {
	should := require.New(t)
	var testData = []struct {
		path           string
		splitPathInfo  string
		scriptFileName string
		root           string
		expect         pathInfo
		expectErr      error
	}{
		{
			path:          "/test.php/foo/bar.php",
			splitPathInfo: `^(.+?\.php)(/.*)$`,
			root:          "/var/www",
			expect: pathInfo{
				ScriptName: "/test.php",
				PathInfo:   "/foo/bar.php",
			},
			expectErr: nil,
		},
		{
			path:          "/test.php/foo/bar.baz",
			splitPathInfo: `^(.+\.php)(.*)$`,
			root:          "/var/www",
			expect: pathInfo{
				ScriptName: "/test.php",
				PathInfo:   "/foo/bar.baz",
			},
			expectErr: nil,
		},
	}
	for idx, tc := range testData {
		r := &Rule{
			Root:          tc.root,
			SplitPathInfo: regexp.MustCompile(tc.splitPathInfo),
		}
		actual, err := r.newPathInfo([]byte(tc.path))
		should.Equal(tc.expectErr, err, "case %d fail", idx)
		should.Equal(tc.expect, actual, "case %d fail", idx)
	}
}

func TestRule_getScriptFileName(t *testing.T) {
}
