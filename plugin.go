package durian

import (
	"github.com/caibirdme/durian/server"
	// plug in directives
	_ "github.com/caibirdme/durian/gzip"
	_ "github.com/caibirdme/durian/header"
	_ "github.com/caibirdme/durian/log"
	_ "github.com/caibirdme/durian/not_found"
	_ "github.com/caibirdme/durian/response"
	_ "github.com/caibirdme/durian/reverse_proxy"
	_ "github.com/caibirdme/durian/rewrite"
	_ "github.com/caibirdme/durian/root"
	_ "github.com/caibirdme/durian/status"
	_ "github.com/caibirdme/durian/timeout"
	"github.com/mholt/caddy"
	"io/ioutil"
	"os"
)

// ReadConfig reads config from given path
func ReadConfig(confPath string) (caddy.Input, error) {
	contents, err := ioutil.ReadFile(confPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return caddy.CaddyfileInput{
		Contents:       contents,
		Filepath:       confPath,
		ServerTypeName: server.FastHTTPServerType,
	}, nil
}

