package middleware

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type StaticOptions struct {
	Enable     bool
	RootPath   string
	Prefix     string
	Exclude    []string
	FilterExts []string
}

//Static 静态文件处理插件
func Static(opts ...*StaticOptions) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		opt := prepareStaticOptions(opts)
		if !opt.Enable {
			ctx.Next()
			return
		}
		if ctx.Request.Method != "GET" && ctx.Request.Method != "HEAD" {
			ctx.Next()
			return
		}
		var rPath = ctx.Request.URL.Path
		//处理特殊文件
		if rPath == "/favicon.ico" || rPath == "/robots.txt" {
			file := path.Join(".", rPath)
			_, err := os.Stat(file)
			if os.IsNotExist(err) {
				ctx.AbortWithStatus(404)
				return
			}
			if err != nil {
				ctx.AbortWithError(500, err)
				return
			}

			ctx.File(file)

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
				ctx.AbortWithStatus(500)
				return
			}
			ctx.AbortWithStatus(404)
			return
		}
		if finfo.IsDir() {
			ctx.AbortWithStatus(404)
			return
		}
		//文件已存在，则返回文件
		ctx.File(fPath)
		return
	}
}

func prepareStaticOptions(options []*StaticOptions) *StaticOptions {
	var opt *StaticOptions
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
	// if len(opt.FilterExts) == 0 {
	// 	opt.FilterExts = append(opt.FilterExts, ".jpg", ".png", ".js", ".css", ".html", ".xml")
	// }
	return opt
}
