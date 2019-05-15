package status

import (
	super "github.com/caibirdme/durian/server"
	"github.com/mholt/caddy"
	"github.com/valyala/fasthttp"
	"strconv"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveStatus, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	cfg, err := parseStatus(c)
	if err != nil {
		return err
	}
	super.GetConfig(c).AddMiddleware(func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if cfg.location.Match(ctx.Path()) {
				ctx.SetStatusCode(cfg.Code)
			}
			next(ctx)
		}
	})
	return nil
}

type StatusConfig struct {
	Code     int
	location super.LocationMatcher
}

func parseStatus(c *caddy.Controller) (*StatusConfig, error) {
	// skip status keyword
	c.Next()

	firstLine := c.RemainingArgs()
	n := len(firstLine)
	location, err := super.NewLocationMatcher(firstLine[:n-1])
	if err != nil {
		return nil, c.Err(err.Error())
	}
	code, err := strconv.Atoi(firstLine[n-1])
	if err != nil {
		return nil, err
	}
	return &StatusConfig{
		location: location,
		Code:     code,
	}, nil
}
