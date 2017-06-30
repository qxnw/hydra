package file

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"path"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/utility"
)

func (s *fileProxy) saveFileFromHTTPRequest(ctx *context.Context) (r string, t int, err error) {
	name, err := ctx.GetInput().Get("name")
	if err != nil {
		err = fmt.Errorf("输入参数input未传入name参数(err:%v)", err)
		return
	}
	root, ok := ctx.GetArgs()["root"]
	if !ok {
		err = fmt.Errorf("输入参数args未配置root目录参数")
		return
	}

	httpRequest, ok := ctx.GetExt()["__func_http_request_"]
	if !ok {
		err = errors.New("未找到__func_http_request_")
		return
	}
	f, ok := httpRequest.(*http.Request)
	if !ok {
		err = errors.New("未找到__func_http_request_类型错误，不是*http.Request")
		return
	}

	uf, _, err := f.FormFile(name)
	if err != nil {
		err = fmt.Errorf("无法读取上传的文件:%s(err:%v)", name, err)
		return
	}
	defer uf.Close()
	name = fmt.Sprintf("%s%s", utility.GetGUID(), path.Ext(name))
	nfilePath := fmt.Sprintf("%s/%s", root, name)
	nf, err := os.Create(nfilePath)
	if err != nil {
		err = fmt.Errorf("保存文件失败:%s(err:%v)", nfilePath, err)
		return
	}
	defer nf.Close()
	io.Copy(nf, uf)
	return name, 200, nil
}
