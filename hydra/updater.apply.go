package hydra

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/qxnw/lib4go/archiver"
	"github.com/qxnw/lib4go/osext"
	"github.com/qxnw/lib4go/security/crc32"
)

type Updater struct {
	targetPath   string
	currentDir   string
	newDir       string
	oldDir       string
	needRollback bool
}

func NewUpdater() (u *Updater, err error) {
	u = &Updater{}
	u.targetPath, err = osext.Executable()
	if err != nil {
		return nil, err
	}
	u.currentDir = filepath.Dir(u.targetPath)
	u.newDir = u.currentDir + ".new"
	u.oldDir = u.currentDir + ".old"
	return
}

//Apply 更新文件
func (u *Updater) Apply(update io.Reader, opts UpdaterOptions) (err error) {
	//读取文件内容
	var newBytes []byte
	if newBytes, err = ioutil.ReadAll(update); err != nil {
		return err
	}
	if opts.CRC32 > 0 {
		if crc32.Encrypt(newBytes) != opts.CRC32 {
			err = fmt.Errorf("文件校验值有误")
			return
		}
	}

	//创建目标目录
	defer os.Chdir(u.currentDir)
	os.RemoveAll(u.newDir)
	err = os.MkdirAll(u.newDir, 0755)
	if err != nil {
		err = fmt.Errorf("权限不足，无法创建文件:%s(err:%v)", u.newDir, err)
		return err
	}
	defer os.RemoveAll(u.newDir)

	//读取归档并解压文件
	archiver := archiver.MatchingFormat(opts.TargetName)
	if archiver == nil {
		err = fmt.Errorf("文件不是有效的归档或压缩文件")
		return
	}
	err = archiver.Read(bytes.NewReader(newBytes), u.newDir)
	if err != nil {
		err = fmt.Errorf("读取归档文件失败:%v", err)
		return
	}
	//备份当前目录
	err = os.RemoveAll(u.oldDir)
	if err != nil {
		err = fmt.Errorf("移除目录失败:%s(err:%v)", u.oldDir, err)
		return
	}
	err = os.Rename(u.currentDir, u.oldDir)
	if err != nil {
		err = fmt.Errorf("无法修改当前工作目录:%s(%s)(err:%v)", u.currentDir, u.oldDir, err)
		return err
	}
	u.needRollback = true
	//将新的目标文件修改为当前目录
	err = os.Rename(u.newDir, u.currentDir)
	if err != nil {
		err = fmt.Errorf("重命名文件夹失败:%v", err)
		return
	}
	err = os.Chdir(u.currentDir)
	if err != nil {
		err = fmt.Errorf("切换工作目录失败:%v", err)
		return
	}
	//err = os.Chtimes(u.currentDir, time.Now(), time.Now())
	//if err != nil {
	//err = fmt.Errorf("修改文件夹时间失败:%v", err)
	//return
	//}
	return
}

//Rollback 回滚当前更新
func (u *Updater) Rollback() error {
	if !u.needRollback {
		return nil
	}
	defer os.Chdir(u.currentDir)
	if _, err := os.Stat(u.oldDir); os.IsNotExist(err) {
		return fmt.Errorf("无法回滚，原备份文件(%s)不存在", u.oldDir)
	}

	//当前工作目录不存在，直接将备份目录更改为当前目录
	if _, err := os.Stat(u.currentDir); os.IsNotExist(err) {
		return os.Rename(u.oldDir, u.currentDir)
	}
	os.RemoveAll(u.currentDir)
	err := os.Rename(u.oldDir, u.currentDir)
	if err != nil {
		return err
	}
	return nil
}

//UpdaterOptions 文件更新选项
type UpdaterOptions struct {
	CRC32      uint32
	TargetName string
}
