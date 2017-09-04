package registry

import (
	"time"

	"fmt"

	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/registry"
	"github.com/qxnw/lib4go/utility"
)

type StandaloneRegistry struct {
	chilren map[string]time.Time
	files   map[string]string
	checker Checker
	done    bool
}

func (l *StandaloneRegistry) SetChecker(c Checker) {
	l.checker = c
}
func (l *StandaloneRegistry) Exists(path string) (bool, error) {
	return l.checker.Exists(path), nil
}
func (l *StandaloneRegistry) WatchChildren(path string) (data chan registry.ChildrenWatcher, err error) {
	data = make(chan registry.ChildrenWatcher, 1)
	go func() {
	START:
		for {
			select {
			case <-time.After(time.Second * 2):
				if l.done {
					break START
				}
				b := l.checker.Exists(path)
				if !b {
					continue
				}

				modify, err := l.checker.LastModeTime(path)
				if err != nil {
					continue
				}
				if t, ok := l.chilren[path]; !ok || t != modify {
					l.chilren[path] = modify
					ve := &valuesEntity{path: path}
					ve.values, ve.Err = l.checker.ReadDir(path)
					data <- ve
				}
			}
		}
	}()

	return data, nil
}
func (l *StandaloneRegistry) Update(path string, data string, version int32) (err error) {
	return l.checker.WriteFile(path, data)
}
func (l *StandaloneRegistry) WatchValue(path string) (data chan registry.ValueWatcher, err error) {
	data = make(chan registry.ValueWatcher, 1)
	go func() {
	START:
		for {
			select {
			case <-time.After(time.Second * 2):
				if l.done {
					break START
				}
				b := l.checker.Exists(path)
				if !b {
					continue
				}
				modify, err := l.checker.LastModeTime(path)
				if err != nil {
					continue
				}
				if t, ok := l.chilren[path]; !ok || t != modify {
					l.chilren[path] = modify
					ve := &valueEntity{path: path}
					ve.Value, ve.Err = l.checker.ReadAll(path)
					data <- ve
				}
			}
		}
	}()

	return data, nil
}
func (l *StandaloneRegistry) GetChildren(path string) (data []string, version int32, err error) {
	data, err = l.checker.ReadDir(path)
	if err != nil {
		return
	}
	modify, err := l.checker.LastModeTime(path)
	if err != nil {
		return
	}
	version = int32(modify.Unix())
	return

}
func (l *StandaloneRegistry) GetValue(path string) (data []byte, version int32, err error) {
	data, err = l.checker.ReadAll(path)
	if err != nil {
		return
	}
	modify, err := l.checker.LastModeTime(path)
	if err != nil {
		return
	}
	version = int32(modify.Unix())
	return
}
func (l *StandaloneRegistry) CreatePersistentNode(path string, data string) (err error) {
	return l.checker.CreateFile(path, data)
}
func (l *StandaloneRegistry) CreateTempNode(path string, data string) (err error) {
	l.files[path] = path
	return l.checker.CreateFile(path, data)
}
func (l *StandaloneRegistry) CreateSeqNode(path string, data string) (rpath string, err error) {
	rpath = fmt.Sprintf("%s_%s", path, utility.GetGUID())
	l.files[rpath] = rpath
	return rpath, l.checker.CreateFile(rpath, data)
}
func (l *StandaloneRegistry) Delete(path string) error {
	delete(l.files, path)
	return l.checker.Delete(path)
}
func (l *StandaloneRegistry) Close() error {
	for _, f := range l.files {
		l.Delete(f)
	}
	return nil
}
func NewLocalRegistry() (r *StandaloneRegistry, err error) {
	r = &StandaloneRegistry{
		files:   make(map[string]string),
		chilren: make(map[string]time.Time),
	}
	r.checker, err = NewChecker()
	return
}
func newLocalRegistryWithChcker(c Checker) (r *StandaloneRegistry, err error) {
	r = &StandaloneRegistry{
		files:   make(map[string]string),
		checker: c,
		chilren: make(map[string]time.Time),
	}
	return
}

type valueEntity struct {
	Value   []byte
	version int32
	path    string
	Err     error
}
type valuesEntity struct {
	values  []string
	version int32
	path    string
	Err     error
}

func (v *valueEntity) GetPath() string {
	return v.path
}

func (v *valueEntity) GetValue() ([]byte, int32) {
	return v.Value, v.version
}
func (v *valueEntity) GetError() error {
	return v.Err
}

func (v *valuesEntity) GetValue() ([]string, int32) {
	return v.values, v.version
}
func (v *valuesEntity) GetError() error {
	return v.Err
}
func (v *valuesEntity) GetPath() string {
	return v.path
}

type localRegistryResolver struct {
}

func (z *localRegistryResolver) Resolve(servers []string, log *logger.Logger) (Registry, error) {
	return NewLocalRegistry()
}

func init() {
	Register("standalone", &localRegistryResolver{})
}
