package fastcgi

import (
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestRule_newPathInfo(t *testing.T) {
	should := require.New(t)
	var testData = []struct {
		path          string
		splitPathInfo string
		prefix        string
		root          string
		expect        pathInfo
		expectErr     error
	}{
		{
			path:          "/test.php/foo/bar.php",
			splitPathInfo: `^(.+?\.php)(/.*)$`,
			root:          "/var/www",
			expect: pathInfo{
				ScriptName:     "/test.php",
				ScriptFileName: "/var/www/test.php",
				PathInfo:       "/foo/bar.php",
			},
			expectErr: nil,
		},
		{
			path:          "/test.php/foo/bar.baz",
			splitPathInfo: `^(.+\.php)(.*)$`,
			root:          "/var/www",
			expect: pathInfo{
				ScriptName:     "/test.php",
				ScriptFileName: "/var/www/test.php",
				PathInfo:       "/foo/bar.baz",
			},
			expectErr: nil,
		},
	}
	for idx, tc := range testData {
		r := &Rule{
			Root:           tc.root,
			FilenamePrefix: tc.prefix,
			SplitPathInfo:  regexp.MustCompile(tc.splitPathInfo),
		}
		actual, err := r.newPathInfo([]byte(tc.path))
		should.Equal(tc.expectErr, err, "case %d fail", idx)
		should.Equal(tc.expect, actual, "case %d fail", idx)
	}
}
