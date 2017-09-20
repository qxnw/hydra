// Copyright 2015 The WebServer Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

type StaticOptions struct {
	Enable     bool
	RootPath   string
	Prefix     string
	Exclude    []string
	FilterExts []string
}

//Static 静态文件处理插件
func Static(opts ...StaticOptions) HandlerFunc {
	return func(ctx *Context) {
		opt := prepareStaticOptions(opts)
		if ctx.Req().Method != "GET" && ctx.Req().Method != "HEAD" {
			ctx.Next()
			return
		}

		var rPath = ctx.Req().URL.Path
		//处理特殊文件
		if rPath == "/favicon.ico" || rPath == "/robots.txt" {
			file := path.Join(".", rPath)
			fi, err := os.Stat(file)
			if os.IsNotExist(err) {
				ctx.WriteHeader(404)
				ctx.Write([]byte("文件不存在"))
				return
			}
			if fi != nil {
				err = ctx.ServeFile(file)
			}
			if err != nil {
				ctx.WriteHeader(500)
				ctx.Write([]byte(err.Error()))
				return
			}
			return
		}

		//检查是否定义静态文件前缀
		if !strings.HasPrefix(rPath, opt.Prefix) {
			ctx.Next()
			return
		}

		rPath = rPath[len(opt.Prefix):]
		dir := filepath.Dir(rPath)

		//检查是否排除特殊名称
		for _, v := range opt.Exclude {
			if strings.Contains(dir, v) {
				ctx.Next()
				return
			}
		}

		fPath, _ := filepath.Abs(filepath.Join(opt.RootPath, rPath))
		//检查文件后缀
		if len(opt.FilterExts) > 0 {
			var matched bool
			for _, ext := range opt.FilterExts {
				if filepath.Ext(fPath) == ext {
					matched = true
					break
				}
			}
			if !matched {
				ctx.Next()
				return
			}
		}

		//检查文件是否存在
		finfo, err := os.Stat(fPath)
		if err != nil {
			if !os.IsNotExist(err) {
				ctx.Result = InternalServerError(err.Error())
				ctx.HandleError()
				return
			}
			ctx.NotFound()
			return
		}
		if finfo.IsDir() {
			ctx.NotFound()
			return
		}
		//文件已存在，则返回文件
		err = ctx.ServeFile(fPath)
		if err != nil {
			ctx.Result = InternalServerError(err.Error())
			ctx.HandleError()
		}
		return
	}
}

func prepareStaticOptions(options []StaticOptions) StaticOptions {
	var opt StaticOptions
	if len(options) > 0 {
		opt = options[0]
	}
	if len(opt.Exclude) == 0 {
		opt.Exclude = append(opt.Exclude, "bin", "conf")
	}
	// Defaults
	if len(opt.RootPath) == 0 {
		opt.RootPath = "../"
	}
	if len(opt.Prefix) == 0 {
		opt.Prefix = "/"
	}
	if len(opt.Prefix) > 0 {
		if opt.Prefix[0] != '/' {
			opt.Prefix = "/" + opt.Prefix
		}
	}
	if len(opt.FilterExts) == 0 {
		opt.FilterExts = append(opt.FilterExts, ".jpg", ".png", ".js", ".css", ".html", ".xml")
	}
	return opt
}
