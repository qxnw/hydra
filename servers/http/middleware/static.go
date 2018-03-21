package middleware

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/conf"
)

func checkExt(s *conf.Static, rPath string) bool {
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
func MustStatic(s *conf.Static, rPath string) (b bool, xname string) {
	if len(rPath) < len(s.Prefix) {
		return checkExt(s, rPath), rPath
	}
	if strings.HasPrefix(rPath, s.Prefix) {
		return true, strings.TrimPrefix(rPath, s.Prefix)
	}
	b = checkExt(s, rPath)
	xname = strings.TrimPrefix(rPath, s.Prefix)
	return
}
func getDefPath(p string) string {
	if p == "" || p == "/" {
		return "index.html"
	}
	return p
}

//Static 静态文件处理插件
func Static(cnf *conf.MetadataConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		opt, ok := cnf.GetMetadata("static").(*conf.Static)
		if !ok || opt.Disable || ctx.Request.Method != "GET" && ctx.Request.Method != "HEAD" {
			ctx.Next()
			return
		}

		var rPath = ctx.Request.URL.Path
		s, xname := MustStatic(opt, rPath)
		if s {
			setExt(ctx, "static")
			fPath, _ := filepath.Abs(filepath.Join(opt.Dir, getDefPath(xname)))
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
