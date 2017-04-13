package standalone

import (
	"io/ioutil"
	"os"
	"time"
)

type checker interface {
	Exists(filename string) bool
	LastModeTime(path string) (t time.Time, err error)
	ReadDir(path string) (r []string, err error)
	ReadAll(path string) (buf []byte, err error)
}

type fileChecker struct {
}

func (f fileChecker) Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
func (f fileChecker) LastModeTime(path string) (t time.Time, err error) {
	p, err := os.Stat(path)
	if err != nil {
		return
	}
	t = p.ModTime()
	return
}
func (f fileChecker) ReadDir(path string) (r []string, err error) {
	children, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	r = make([]string, 0, len(children))
	for _, v := range children {
		r = append(r, v.Name())
	}
	return
}
func (f fileChecker) ReadAll(path string) (buf []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}
