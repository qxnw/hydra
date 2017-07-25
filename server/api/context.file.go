package api

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"time"

	"fmt"

	"github.com/qxnw/lib4go/concurrent/cmap"
)

var staticFileCache cmap.ConcurrentMap

func init() {
	staticFileCache = cmap.New(16)
}

type staticHolder struct {
	*bytes.Reader

	modTime  time.Time
	size     int64
	encoding string
}

func isOk(s *staticHolder, fi os.FileInfo) bool {
	if s == nil {
		return false
	}
	return s.modTime == fi.ModTime() && s.size == fi.Size()
}

func (ctx *Context) ServeFile(path string) error {

	b, buff, err := staticFileCache.SetIfAbsentCb(path, func(input ...interface{}) (interface{}, error) {
		path := input[0].(string)
		fi, err := os.Open(path)
		if err != nil {
			msg, code := toHTTPError(err)
			http.Error(ctx, msg, code)
			return nil, err
		}
		defer fi.Close()

		d, err := fi.Stat()
		if err != nil {
			msg, code := toHTTPError(err)
			http.Error(ctx, msg, code)
			return nil, err
		}

		if d.IsDir() {
			http.Error(ctx, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return nil, fmt.Errorf(http.StatusText(http.StatusForbidden))
		}
		var bufferWriter bytes.Buffer
		_, err = io.Copy(&bufferWriter, fi)
		if err != nil {
			return nil, err
		}
		holder := &staticHolder{Reader: bytes.NewReader(bufferWriter.Bytes()), modTime: d.ModTime(), size: int64(bufferWriter.Len())}
		return holder, nil
	}, path)
	if err != nil {
		return err
	}
	holder := buff.(*staticHolder)
	if !b {
		if fi, _ := os.Stat(path); fi != nil && !isOk(holder, fi) {
			staticFileCache.Remove(path)
			holder = nil
			return ctx.ServeFile(path)
		}
	}
	http.ServeContent(ctx, ctx.Req(), path, holder.modTime, holder.Reader)
	return nil
}
