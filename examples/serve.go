package main

import (
	"io/ioutil"
	"log"
	"os"

	_ "github.com/caibirdme/caddy-fasthttp"
	"github.com/mholt/caddy"
)

func init() {
	caddy.TrapSignals()
	// configure default caddyfile
	caddy.SetDefaultCaddyfileLoader("default", caddy.LoaderFunc(defaultLoader))
}

func main() {
	caddy.AppName = "haha"
	caddy.AppVersion = "0.0.1"

	// load caddyfile
	caddyfile, err := caddy.LoadCaddyfile("fasthttp")
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(os.Stdout)
	// start caddy server
	instance, err := caddy.Start(caddyfile)
	if err != nil {
		log.Fatal(err)
	}

	instance.Wait()
}

// provide loader function
func defaultLoader(serverType string) (caddy.Input, error) {
	contents, err := ioutil.ReadFile(caddy.DefaultConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return caddy.CaddyfileInput{
		Contents:       contents,
		Filepath:       caddy.DefaultConfigFile,
		ServerTypeName: serverType,
	}, nil
}
