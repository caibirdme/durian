package rewrite

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"regexp"
	"strconv"
)

//Rewriter ...
type URLRewriter struct {
	from *regexp.Regexp
	to string
}

func (u *URLRewriter) Rewrite(path []byte) ([]byte, bool) {
	res := u.from.FindSubmatch(path)
	if len(res) == 0 {
		return nil, false
	}
	count := len(res)-1;
	ans := []byte(u.to)
	var dollar = []byte("{")
	for i:=1; i<=count; i++ {
		t := strconv.AppendInt(dollar[:1], int64(i), 10)
		t = append(t, '}')
		ans = bytes.ReplaceAll(ans, t, res[i])
	}
	return ans, true
}

func (u *URLRewriter) Handle(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		newPath,ok := u.Rewrite(ctx.Path())
		if ok && len(newPath) > 0 {
			ctx.URI().SetPathBytes(newPath)
		}
		next(ctx)
	}
}

func NewRewriter(from string, to string) (*URLRewriter,error) {
	re,err := regexp.Compile(from)
	if err != nil {
		return nil, err
	}
	return &URLRewriter{from: re, to: to},nil
}

