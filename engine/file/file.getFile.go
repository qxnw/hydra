package file

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"path"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/utility"
)

func (s *fileProxy) getParams(ctx *context.Context) (input transform.ITransformGetter, args map[string]string, err error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil || ctx.Input.Params == nil {
		err = fmt.Errorf("input,params,args不能为空:%v", ctx.Input)
		return
	}
	input, ok := ctx.Input.Input.(transform.ITransformGetter)
	if !ok {
		err = fmt.Errorf("输入参数input不是transform.ITransformGetter类型")
		return
	}
	args, ok = ctx.Input.Args.(map[string]string)
	if !ok {
		err = fmt.Errorf("输入参数args不是map[string]string类型")
		return
	}
	return
}
func (s *fileProxy) saveFileFromHTTPRequest(ctx *context.Context) (r string, err error) {
	input, args, err := s.getParams(ctx)
	if err != nil {
		return
	}
	name, err := input.Get("name")
	if err != nil {
		err = fmt.Errorf("输入参数input未传入name参数(err:%v)", err)
		return
	}
	root, ok := args["root"]
	if !ok {
		err = fmt.Errorf("输入参数args未配置root参数")
		return
	}

	httpRequest := ctx.Ext["__func_http_request_"]
	if httpRequest == nil {
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
		err = fmt.Errorf("读取文件错误:%s(err:%v)", name, err)
		return
	}
	defer uf.Close()
	nfilePath := fmt.Sprintf("%s/%s%s", root, utility.GetGUID(), path.Ext(name))
	nf, err := os.Create(nfilePath)
	if err != nil {
		err = fmt.Errorf("创建文件失败:%s(err:%v)", nfilePath, err)
		return
	}
	defer nf.Close()
	io.Copy(nf, uf)
	return nfilePath, nil
}
