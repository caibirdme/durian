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

func Test_convertHeader2EnvParam(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "one -",
			args: "Content-Type",
			want: "CONTENT_TYPE",
		},
		{
			name: "two -",
			args: "Foo-Bar-Baz",
			want: "FOO_BAR_BAZ",
		},
		{
			name: "zero -",
			args: "FooBarBaz",
			want: "FOOBARBAZ",
		},
		{
			name: "empty",
			args: "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertHeader2EnvParam([]byte(tt.args)); string(got) != tt.want {
				t.Errorf("convertHeader2EnvParam() = %s, want %v", got, tt.want)
			}
		})
	}
}
