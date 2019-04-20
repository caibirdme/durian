package rewrite

import (
	super "github.com/caibirdme/durian/server"
	"github.com/mholt/caddy"
	"strings"
)

func init() {
	caddy.RegisterPlugin(super.DirectiveRewrite, caddy.Plugin{
		ServerType: super.FastHTTPServerType,
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	rule, err := parseRewrite(c)
	if nil != err {
		return err
	}
	r, err := NewRewriter(rule.From, rule.To)
	if err != nil {
		return err
	}
	super.GetConfig(c).AddMiddleware(r.Handle)
	return nil
}

type RewriteRule struct {
	From string
	To   string
}

var nilRule = RewriteRule{}

func parseRewrite(c *caddy.Controller) (RewriteRule, error) {
	// skip rewrite keyword
	c.Next()
	if !c.NextArg() {
		return nilRule, c.ArgErr()
	}
	rule := RewriteRule{From: c.Val()}
	for c.NextBlock() {
		kind := c.Val()
		err := parseKind(kind, c, &rule)
		if nil != err {
			return nilRule, err
		}
	}
	return rule, nil
}

func parseKind(kind string, c *caddy.Controller, rule *RewriteRule) error {
	switch strings.ToLower(kind) {
	case "to":
		if !c.NextArg() {
			return c.ArgErr()
		}
		rule.To = c.Val()
	default:
	}
	return nil
}
