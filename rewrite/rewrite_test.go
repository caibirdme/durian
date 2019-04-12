package rewrite

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestURLRewriter_Rewrite(t *testing.T) {
	var testCases = []struct {
		from string
		to string
		cases [][2]string
	}{
		{
			from: `/foo/\w+/(.*)/qq`,
			to: "/tmp/$1/some",
			cases: [][2]string{
				[2]string{"/foo/wqe/tony/qq", "/tmp/tony/some"},
				{"/foo/wqe/john/qq", "/tmp/john/some"},
			},
		},
		{
			from: `/a/(123.*)/(.*)`,
			to: "/$2/$1",
			cases: [][2]string{
				[2]string{"/a/12345/hello", "/hello/12345"},
				{"/a/123come_on/foobar", "/foobar/123come_on"},
			},
		},
		{
			from: `/a/(1.*)/(.*)`,
			to: "/$2/$1/$1_$2",
			cases: [][2]string{
				[2]string{"/a/12/h", "/h/12/12_h"},
			},
		},
	}
	for _,tc := range testCases {
		t.Run(tc.from, func(tt *testing.T) {
			should := require.New(tt)
			r,err := NewRewriter(tc.from, tc.to)
			should.NoError(err)
			for _,pair := range tc.cases {
				actual,ok := r.Rewrite([]byte(pair[0]))
				should.True(ok)
				should.Equal(string(pair[1]), string(actual))
			}
		})
	}
}