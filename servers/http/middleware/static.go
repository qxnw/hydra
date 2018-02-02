package middleware

import (
	"fmt"
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

//MustStatic 判断当前文件是否一定是静态文件 0:非静态文件  1：是静态文件  2：未知
func (s *StaticOptions) MustStatic(rPath string) int {
	if strings.HasPrefix(rPath, s.Prefix) {
		return 1
	}
	rPath = rPath[len(s.Prefix):]
	dir := filepath.Dir(rPath)
	//检查是否排除特殊名称
	for _, v := range s.Exclude {
		if strings.Contains(dir, v) {
			return 0
		}
	}
	if len(s.FilterExts) > 0 {
		for _, ext := range s.FilterExts {
			if filepath.Ext(rPath) == ext {
				return 1
			}
		}
		return 0
	}
	return 2
}

//Prepare 准备初始参数
func (s *StaticOptions) Prepare() {
	if len(s.Exclude) == 0 {
		s.Exclude = append(s.Exclude, "bin", "conf")
	}
	if s.RootPath == "" {
		s.RootPath = "../"
	}
	if s.Prefix == "" {
		s.Prefix = "/"
	}
	if s.Prefix[0] != '/' {
		s.Prefix = "/" + s.Prefix
	}
}

//Static 静态文件处理插件
func Static(opt *StaticOptions) gin.HandlerFunc {
	return func(ctx *gin.Context) {
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
				ctx.AbortWithError(500, fmt.Errorf("%s,err:%v", rPath, err))
				return
			}
			ctx.File(file)
			return
		}
		if !opt.Enable {
			ctx.Next()
			return
		}
		s := opt.MustStatic(rPath)
		switch s {
		case 0:
			ctx.Next()
			return
		case 1:
			fPath, _ := filepath.Abs(filepath.Join(opt.RootPath, rPath[len(opt.Prefix):]))
			finfo, err := os.Stat(fPath)
			if err != nil {
				if !os.IsNotExist(err) {
					ctx.AbortWithError(500, fmt.Errorf("%s,err:%v", fPath, err))
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
		case 2:
			fPath, _ := filepath.Abs(filepath.Join(opt.RootPath, rPath))
			finfo, err := os.Stat(fPath)
			if err != nil {
				if !os.IsNotExist(err) {
					ctx.AbortWithError(500, fmt.Errorf("%s,err:%v", fPath, err))
					return
				}
				ctx.Next()
				return
			}
			if finfo.IsDir() {
				ctx.Next()
				return
			}
			//文件已存在，则返回文件
			ctx.File(fPath)
			return
		}
	}
}
