package hydra

import (
	"bytes"
	"crypto"
	"fmt"
	"net/http"

	"github.com/zkfy/go-update"
)

//update 下载安装包并解压到临时目录，停止所有服务器，并拷贝到当前工作目录
func updateNow(url string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var buff []byte
	_, err = resp.Body.Read(buff)
	if err != nil {
		return
	}
	err = update.Apply(bytes.NewReader(buff), update.Options{
		Hash:     crypto.SHA256,
		Checksum: buff,
	})
	if err != nil {
		if err1 := update.RollbackError(err); err1 != nil {
			err = fmt.Errorf("%+v,%v", err, err1)
			return
		}
	}
	return err
}
