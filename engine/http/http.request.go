package http

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net/http"
)

func (s *httpProxy) httpRequest(name string, mode string, service string, ctx *context.Context) (response *context.WebResponse, err error) {

	charset := ctx.Input.GetArgsValue("charset", "utf-8")
	method := strings.ToUpper(ctx.Input.GetArgsValue("method", "POST"))
	host, err := ctx.Input.GetArgsByName("host")
	if err != nil {
		return
	}
	u, err := url.Parse(path.Join(host, service))
	if err != nil {
		return
	}
	values := u.Query()
	ctx.Input.Input.Each(func(k, v string) {
		values.Add(k, v)
	})
	header := make(map[string]string)
	header["Cookie"] = fmt.Sprintf("hydra_sid=%s", ctx.Input.Ext["hydra_sid"])
	client := http.NewHTTPClient()
	url := ""
	value := ""
	switch method {
	case "POST":
		url = u.Path
		value = values.Encode()
	default:
		url = fmt.Sprintf("%s?%s", url, values.Encode())
		value = ""
	}
	hc, t, err := client.Request(method, url, value, charset, header)
	if err != nil || t != 200 {
		response.SetStatus(t)
		return
	}
	for _, cookie := range client.Response.Cookies() {
		for _, name := range s.names {
			if name == cookie.Name {
				response.SetParams(name, cookie.Value)
			}
		}
	}
	for _, name := range s.names {
		value := client.Response.Header.Get(name)
		response.SetParams(name, value)
	}
	response.Success(hc)
	return
}
