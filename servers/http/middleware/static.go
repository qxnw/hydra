package middleware

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type StaticOptions struct {
	Enable   bool
	RootPath string
	Prefix   string
	Exclude  []string
	Exts     []string
}

func (s *StaticOptions) checkExt(rPath string) bool {
	name := filepath.Base(rPath)
	for _, v := range s.Exclude {
		if name == v {
			return true
		}
	}
	hasExt := strings.Contains(filepath.Ext(name), ".")
	if len(s.Exts) > 0 {
		if s.Exts[0] == "*" && (rPath == "/" || hasExt) {
			return true
		}
		pExt := filepath.Ext(name)
		for _, ext := range s.Exts {
			if pExt == ext {
				return true
			}
		}
	}
	return false
}

//MustStatic 判断当前文件是否一定是静态文件 0:非静态文件  1：是静态文件  2：未知
func (s *StaticOptions) MustStatic(rPath string) (b bool, xname string) {
	if len(rPath) < len(s.Prefix) {
		return s.checkExt(rPath), rPath
	}
	if strings.HasPrefix(rPath, s.Prefix) {
		return true, strings.TrimPrefix(rPath, s.Prefix)
	}
	b = s.checkExt(rPath)
	xname = strings.TrimPrefix(rPath, s.Prefix)
	return
}
func (s *StaticOptions) getDefPath(p string) string {
	if p == "" || p == "/" {
		return "index.html"
	}
	return p
}

//Static 静态文件处理插件
func Static(opt *StaticOptions) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !opt.Enable || ctx.Request.Method != "GET" && ctx.Request.Method != "HEAD" {
			ctx.Next()
			return
		}
		var rPath = ctx.Request.URL.Path
		s, xname := opt.MustStatic(rPath)
		if s {
			setExt(ctx, "static")
			fPath, _ := filepath.Abs(filepath.Join(opt.RootPath, opt.getDefPath(xname)))
			finfo, err := os.Stat(fPath)
			if err != nil {
				if os.IsNotExist(err) {
					ctx.AbortWithError(404, fmt.Errorf("找不到文件:%s", fPath))
					return
				}
				ctx.AbortWithError(500, fmt.Errorf("%s,err:%v", fPath, err))
				return
			}
			if finfo.IsDir() {
				ctx.AbortWithError(404, fmt.Errorf("找不到文件:%s", fPath))
				return
			}
			//文件已存在，则返回文件
			ctx.File(fPath)
			ctx.Abort()
			return
		}
		ctx.Next()
		return

	}
}
