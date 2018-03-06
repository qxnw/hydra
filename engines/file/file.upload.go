package file

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/security/crc32"
	"github.com/qxnw/lib4go/utility"
)

//FileUpload 文件上传服务
func FileUpload() component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetMapResponse()
		name, err = ctx.Request.Form.Get("name")
		if err != nil {
			err = fmt.Errorf("输入参数input未传入name参数(err:%v)", err)
			return
		}
		root, err := ctx.Request.Setting.Get("root")
		if err != nil {
			return
		}
		f, err := ctx.Request.Http.Get()
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
			response.SetContent(500, err)
			return
		}
		buff, err := ioutil.ReadAll(nf)
		if err != nil {
			err = fmt.Errorf("读取文件失败:%v", err)
			return
		}
		crc := crc32.Encrypt(buff)
		response.SetContent(200, map[string]interface{}{
			"name": name,
			"crc":  crc,
		})
		return
	}
}