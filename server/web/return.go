// Copyright 2015 The WebServer Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server/api"
)

type StatusResult struct {
	Code   int
	Result interface{}
	Type   int
}

func isNil(a interface{}) bool {
	if a == nil {
		return true
	}
	aa := reflect.ValueOf(a)
	return !aa.IsValid() || (aa.Type().Kind() == reflect.Ptr && aa.IsNil())
}

func (w *WebServer) Return() api.HandlerFunc {
	return func(ctx *api.Context) {
		action := ctx.Action()
		ctx.Next()

		// if no route match or has been write, then return
		if action == nil || ctx.Written() {
			return
		}

		// if there is no return value or return nil
		if isNil(ctx.Result) {
			// then we return blank page
			ctx.Result = ""
		}

		if len(ctx.Server.Headers) > 0 {
			for k, v := range ctx.Server.Headers {
				ctx.Header().Set(k, v)
			}
		}
		switch ctx.Result.(type) {
		case error:
			viewPath := fmt.Sprintf("%s/%s%s", w.viewRoot, w.errorView, w.viewExt)
			err := w.viewTmpl.Execute(ctx.ResponseWriter, viewPath, ctx.Result)
			if err != nil {
				ctx.Errorf("web.response.error: %v", err)
			}
		case *context.Response:
			response := ctx.Result.(*context.Response)
			if response.IsRedirect() {
				return
			}
			view, ok := response.Params["__view"]
			if ok && view == "NONE" {
				write(ctx, response)
				return
			}
			if len(ctx.Header().Get("Content-Type")) <= 0 {
				ctx.Header().Set("Content-Type", "text/html; charset=UTF-8")
			}
			if view == nil || view.(string) == "" {
				view = ctx.ServiceName
			}
			viewPath := fmt.Sprintf("%s%s%s", w.viewRoot, view, w.viewExt)
			err := w.viewTmpl.Execute(ctx.ResponseWriter, viewPath, response.Content)
			if err != nil {
				ctx.Errorf("web.response.error: %v", err)
			}
		default:
			ctx.WriteHeader(505)
			ctx.Write([]byte("系统错误"))
		}

	}
}

func write(ctx *api.Context, response *context.Response) {
	rt := api.JsonResponse
	if tp, ok := response.Params["Content-Type"].(string); ok {
		if strings.Contains(tp, "xml") {
			rt = api.XmlResponse
		} else if strings.Contains(tp, "json") {
			rt = api.JsonResponse
		} else {
			rt = api.AutoResponse
		}
	}

	result := ctx.Result
	if rt == api.JsonResponse {
		encoder := json.NewEncoder(ctx)
		if len(ctx.Header().Get("Content-Type")) <= 0 {
			ctx.Header().Set("Content-Type", "application/json; charset=UTF-8")
		}

		switch res := result.(type) {
		case error:
			if response.Status == 0 {
				response.Status = http.StatusInternalServerError
			}
			ctx.WriteHeader(response.Status)
			encoder.Encode(map[string]string{
				"err": res.Error(),
			})
		case string:
			if response.Status == 0 {
				response.Status = http.StatusOK
			}
			ctx.WriteHeader(response.Status)
			ctx.WriteString(res)
		case json.RawMessage:
			if response.Status == 0 {
				response.Status = http.StatusOK
			}
			ctx.WriteHeader(response.Status)
			encoder.Encode(res)
		case []byte:
			if response.Status == 0 {
				response.Status = http.StatusOK
			}
			ctx.WriteHeader(response.Status)
			ctx.Write(res)
		default:
			if response.Status == 0 {
				response.Status = http.StatusOK
			}
			ctx.WriteHeader(response.Status)
			if result == nil {
				return
			}
			err := encoder.Encode(result)
			if err != nil {
				ctx.Result = err
				encoder.Encode(map[string]string{
					"err": err.Error(),
				})
			}
		}

		return
	} else if rt == api.XmlResponse {
		encoder := xml.NewEncoder(ctx)
		if len(ctx.Header().Get("Content-Type")) <= 0 {
			ctx.Header().Set("Content-Type", "application/xml; charset=UTF-8")
		}
		switch res := result.(type) {
		case error:
			ctx.WriteHeader(response.Status)
			encoder.Encode(XmlError{
				Content: res.Error(),
			})
		case string:
			ctx.WriteHeader(response.Status)
			ctx.WriteString(res)
		case []byte:
			ctx.WriteHeader(response.Status)
			ctx.Write(res)
		default:
			ctx.WriteHeader(response.Status)
			if result == nil {
				return
			}
			err := encoder.Encode(result)
			if err != nil {
				ctx.Result = err
				encoder.Encode(XmlError{
					Content: err.Error(),
				})
			}
		}
		return
	}
	if len(ctx.Header().Get("Content-Type")) <= 0 {
		ctx.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	}
	switch res := result.(type) {
	case error:
		ctx.HandleError()
		ctx.WriteHeader(response.Status)
		ctx.WriteString(fmt.Sprintf("%v", res))
	case []byte:
		ctx.WriteHeader(response.Status)
		ctx.Write(res)
	case string:
		ctx.WriteHeader(response.Status)
		ctx.WriteString(res)
	default:
		ctx.WriteHeader(response.Status)
		if result == nil {
			return
		}
		ctx.WriteString(fmt.Sprintf("%v", res))
	}
}

type XmlError struct {
	XMLName xml.Name `xml:"err"`
	Content string   `xml:"content"`
}
