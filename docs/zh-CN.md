# Durian

Durian（榴莲）是一个通用的web server，它的目标是成为apache或者nginx的替代（当然目前还替代不了）。
另一方面，durian同时也是一个模块化、插件化的开发框架，用户可以基于Durian开发自己的业务逻辑。

Durian本身已经是一个功能比较齐全且性能强大的web server，能够通过简单的配置完成用户需求，比如：
```
:8051 {

    # url命中pattern之后直接输出
    response {
        pattern /foo/\d+/.*
        body "{\"name\": \"caibirdme\", \"age\": 25, \"some\":[12,3,4]}"
        content_type Application/json
    }
    
    # 反向代理
    proxy {
        pattern /bar/.*
        upstream {
            10.10.10.1
            10.10.10.2
        }
    }

    # 配置日志
    log {
        access_path ./access.log
        err_path ./error.log
        # 配置日志输出内容，具体可以输出哪些可以参见文档
        # 这里也可以只写一个format后面啥都不写，表示使用默认的格式
        format
    }
}

```
通过以上配置文件，可以启动一个监听8051端口的服务：
* 对/foo/123/xx 这样的请求，直接返回一段json
* 把/bar/xx 这样的请求，等概率地转发到upstream的这两台机器上
access.log里每条日志都是json（基于zap）：
```json
{"remote_addr":"127.0.0.1:50928","host":"localhost:8051","method":"GET","request_uri":"/foo/123/bar","status":200,"start_time":"2019-04-27T19:57:06.733+0800","process_time":"12.34µs","bytes_sent":155,"user_agent":"curl/7.58.0","response_body":"{\"name\": \"caibirdme\", \"age\": 25, \"some\":[12,3,4]}"}
```

使用Durian作为web server有以下优势：
* 配置简单（分分钟看明白）
* 高性能（net/http的5倍以上，在我的烂机器上测的）
* 热重启（0宕机时间）
* 热更新（即使更新了二进制，也0宕机时间）
* 跨平台（Windows Mac Linux都能用，理论上是这样）
* Pidfile可以支持supervisor托管
* 扩展性强（这个是核心卖点）

## 扩展
就像OpenResty一样，依托于Nginx自身的能力，通过Lua去做一些额外的工作。

Durian也可以这样，只需要几行代码，你就可以让你的业务代码运行在Durian之上，在获得高性能同时也能获得了graceful restart的能力。

以下是一个完整的可运行的代码用例：
```go
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
```

* 在init中执行`caddy.TrapSignals()`，各种热重启热更新就设置好了。
* 设置一下应用的Name和Version，你可以乱填
* 调用`router.RegisterPlugin(handler)`注册你自己的代码逻辑
* 读取配置文件并启动服务器
* 等着所有服务都graceful shutdown

最关键的就是注册你的代码逻辑。上面示例中handler签名为

`func(cfg router.RouterConfig) (fasthttp.RequestHandler, error)`

RouterConfig其实就是你在Caddyfile中配置的router那个模块，比如：
```
:8051 {

    router {
        config /home/my/app.toml
    }

    # 配置日志
    log {
        access_path ./access.log
        err_path ./error.log
        format
    }
}
```
此时`cfg.ConfigPath == "/home/my/app.toml"`，这样你就可以读取你自己的配置文件，然后初始化你自己的资源了。你可以用任何基于fasthttp的框架、middleware等等，详细示例可以看examples