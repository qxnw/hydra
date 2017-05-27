package registry

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type Checker interface {
	Exists(filename string) bool
	CreateFile(fileName string, data string) error
	WriteFile(fileName string, data string) error
	LastModeTime(path string) (t time.Time, err error)
	ReadDir(path string) (r []string, err error)
	ReadAll(path string) (buf []byte, err error)
	Delete(path string) error
}

type fileChecker struct {
}

func NewChecker() (Checker, error) {
	if err := checkPrivileges(); err != nil {
		return nil, err
	}
	return &fileChecker{}, nil
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
func (f fileChecker) CreateFile(fileName string, data string) error {
	err := os.MkdirAll(filepath.Dir(fileName), 0666)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, []byte(data), 0666)
}
func (f fileChecker) WriteFile(fileName string, data string) error {
	return ioutil.WriteFile(fileName, []byte(data), 0666)
}
func (f fileChecker) Delete(fileName string) error {
	return os.Remove(fileName)
}
