package main

import (
	"fmt"
	"github.com/buaazp/fasthttprouter"
	"github.com/caibirdme/durian"
	"github.com/caibirdme/durian/router"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"log"
)

func init() {
	caddy.TrapSignals()
}

func main() {
	caddy.AppName = "haha"
	caddy.AppVersion = "0.0.1"
	// 注册我们自己的逻辑
	router.RegisterPlugin(handler)

	// 读取配置
	input, err := durian.ReadConfig("./Caddyfile")

	// 启动服务
	instance, err := caddy.Start(input)
	if err != nil {
		log.Fatal(err)
	}

	// 等待instance下所有server stop
	instance.Wait()
}

// 你自己的业务逻辑，这个里就是你自己通常代码里的main
func handler(cfg router.RouterConfig) (fasthttp.RequestHandler, error) {
	/*
		可以在Caddyfile中指定业务配置文件的路径，之后就能从cfg参数里取到
		router {
			config /home/tom/cfg.toml
		}
	*/

	// your config file path, you can decode it
	_ = cfg.CfgPath

	// 可以使用任何基于fasthttp的router或者框架
	r := fasthttprouter.New()
	// 配置路由
	r.GET("/user/:name", getUserName)

	return r.Handler, nil
}

func getUserName(ctx *fasthttp.RequestCtx) {
	ctx.WriteString(fmt.Sprintf("Hello %v\n", ctx.UserValue("name")))
}
