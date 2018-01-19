package file

import (
	"io/ioutil"
	"fmt"
	"io"
	"os"

	"path"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/utility"
	"github.com/qxnw/lib4go/security/crc32"
)

func (s *fileProxy) saveFileFromHTTPRequest(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	name, err = ctx.Input.Get("name")
	if err != nil {
		err = fmt.Errorf("输入参数input未传入name参数(err:%v)", err)
		return
	}
	root, err := ctx.Input.GetArgsByName("root")
	if err != nil {
		return
	}
	f, err := ctx.HTTP.GetHTTPRequest()
	if err != nil {
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
	_, err = io.Copy(nf, uf)
	if err != nil {
		response.SetStatus(500)
		return
	}
	response.Success(name)
	return
}

func (s *fileProxy) saveFileFromHTTPRequestV2(name string, mode string, service string, ctx *context.Context) (response *context.MapResponse, err error) {
	response = context.GetMapResponse()
	name, err = ctx.Input.Get("name")
	if err != nil {
		err = fmt.Errorf("输入参数input未传入name参数(err:%v)", err)
		return
	}
	root, err := ctx.Input.GetArgsByName("root")
	if err != nil {
		return
	}
	f, err := ctx.HTTP.GetHTTPRequest()
	if err != nil {
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
	_, err = io.Copy(nf, uf)
	if err != nil {
		response.SetStatus(500)
		return
	}
	buff, err := ioutil.ReadAll(nf)
	if err != nil {
		err = fmt.Errorf("读取文件失败:%v", err)
		return
	}
	crc:=crc32.Encrypt(buff)	
	response.Success(map[string]interface{}{
		"name":name,
		"crc":crc,
	})
	return
}
func (s *fileProxy) saveFileFromHTTPRequest2(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	name, err = ctx.Input.Get("name")
	if err != nil {
		return
	}
	root, err := ctx.Input.GetArgsByName("root")
	if err != nil {
		return
	}
	f, err := ctx.HTTP.GetHTTPRequest()
	if err != nil {
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
	_, err = io.Copy(nf, uf)
	if err != nil {
		response.SetStatus(500)
		return
	}
	response.Success(name)
	return
}
