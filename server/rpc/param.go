// Copyright 2015 The WebServer Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import "github.com/qxnw/hydra/context"

func (ctx *Context) Param(key string, defaults ...string) string {
	return ctx.Params().MustString(key, defaults...)
}

func (ctx *Context) ParamStrings(key string, defaults ...[]string) []string {
	return ctx.Params().MustStrings(key, defaults...)
}

func (ctx *Context) ParamEscape(key string, defaults ...string) string {
	return ctx.Params().MustEscape(key, defaults...)
}

func (ctx *Context) ParamInt(key string, defaults ...int) int {
	return ctx.Params().MustInt(key, defaults...)
}

func (ctx *Context) ParamInt32(key string, defaults ...int32) int32 {
	return ctx.Params().MustInt32(key, defaults...)
}

func (ctx *Context) ParamInt64(key string, defaults ...int64) int64 {
	return ctx.Params().MustInt64(key, defaults...)
}

func (ctx *Context) ParamUint(key string, defaults ...uint) uint {
	return ctx.Params().MustUint(key, defaults...)
}

func (ctx *Context) ParamUint32(key string, defaults ...uint32) uint32 {
	return ctx.Params().MustUint32(key, defaults...)
}

func (ctx *Context) ParamUint64(key string, defaults ...uint64) uint64 {
	return ctx.Params().MustUint64(key, defaults...)
}

func (ctx *Context) ParamFloat32(key string, defaults ...float32) float32 {
	return ctx.Params().MustFloat32(key, defaults...)
}

func (ctx *Context) ParamFloat64(key string, defaults ...float64) float64 {
	return ctx.Params().MustFloat64(key, defaults...)
}

func (ctx *Context) ParamBool(key string, defaults ...bool) bool {
	return ctx.Params().MustBool(key, defaults...)
}

type Paramer interface {
	SetParams([]context.Param)
}

func Param() HandlerFunc {
	return func(ctx *Context) {
		if action := ctx.Action(); action != nil {
			if p, ok := action.(Paramer); ok {
				p.SetParams(*ctx.Params())
			}
		}
		ctx.Next()
	}
}
