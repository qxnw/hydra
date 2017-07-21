// Copyright 2015 The WebServer Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"fmt"
	"reflect"

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
		viewPath := fmt.Sprintf("%s%s%s", w.viewRoot, ctx.ServiceName, w.viewExt)
		err := w.viewTmpl.Execute(ctx.ResponseWriter, viewPath, ctx.Result)
		if err != nil {
			ctx.Errorf("web.response.error: %v", err)
		}
	}
}
