package server

import (
	"testing"

	"github.com/mholt/caddy/caddyfile"
	"github.com/stretchr/testify/require"
)

func TestConvertCaddyfile(t *testing.T) {
	var cfg = `
	:8080,:8051 {
		proxy /foo/(\w)/(.*) localhost:8079 {
			policy random
			timeout 1s
		}
	}
	http://foo.com:8081 {
		proxy /foo/(\w)/(.*) localhost:9931 {
			policy random
			timeout 5s
		}
	}
	`
	after, err := caddyfile.ToJSON([]byte(cfg))
	should := require.New(t)
	should.NoError(err)
	should.FailNow("test", "%s", after)
}
