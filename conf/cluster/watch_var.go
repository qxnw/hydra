package cluster

/*
import (
	"fmt"

	"time"

	rs "github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/registry"
)

type valueData struct {
	data    []byte
	version int32
}

var varDirs = []string{"setting", "db", "cache", "mq"}

type watchVarConf struct {
	domain         string
	cache          cmap.ConcurrentMap
	watchValueChan chan string
	dirPaths       []string
	watchDirChan   chan string
	valuePaths     []string
	registry       rs.Registry
}

func newWatchVarConf(domain string) *watchVarConf {
	r := &watchVarConf{domain: domain,
		cache:      cmap.New(2),
		dirPaths:   make([]string, 10),
		valuePaths: make([]string, 10),
	}
	r.watchDirChan = make(chan string, len(varDirs))
	for _, v := range varDirs {
		r.watchDirChan <- fmt.Sprintf("%s/var/%s", domain, v)
	}
	go r.watchPath()
	go r.watchValue()
	return r
}
func (w *watchVarConf) GetValue(path string) ([]byte, error) {
	_, value, err := w.cache.SetIfAbsentCb(path, func(input ...interface{}) (interface{}, error) {
		p := input[0].(string)
		w.watchValueChan <- p
		buffer, v, err := w.registry.GetValue(p)
		if err != nil {
			return nil, err
		}
		return &valueData{data: buffer, version: v}, nil
	})
	if err != nil {
		return []byte{}, err
	}
	return value.(valueData).data, nil
}
func (w *watchVarConf) watchPath() {
	childrenChangedChan := make(chan chan registry.ChildrenWatcher, 10)
	for {
		select {
		case p := <-w.watchDirChan:
			v, err := w.registry.WatchChildren(p)
			if err != nil {
				time.Sleep(time.Second * 10)
				w.watchDirChan <- p
				break
			}
			childrenChangedChan <- v
		case ccc := <-childrenChangedChan:
			select {
			case children := <-ccc:
				if err := children.GetError(); err != nil {
					time.Sleep(time.Second * 10)
					w.watchDirChan <- children.GetPath()
					break
				}
				paths, _ := children.GetValue()
				for _, path := range paths {
					w.watchValueChan <- fmt.Sprintf("%s/%s", children.GetPath(), path)
				}
				w.watchDirChan <- children.GetPath()
			default:
			}
		}
	}
}
func (w *watchVarConf) watchValue() {
	valueChangedChan := make(chan chan registry.ValueWatcher, 10)
	for _, v := range varDirs {
		path := fmt.Sprintf("%s/var/%s", w.domain, v)
		buff, version, err := w.registry.GetValue(path)
		if err != nil {
			continue
		}
		w.cache.Set(path, &valueData{data: buff, version: version})
	}

	for {
		select {
		case p := <-w.watchValueChan:
			v, err := w.registry.WatchValue(p)
			if err != nil {
				time.Sleep(time.Second * 10)
				w.watchValueChan <- p
				break
			}
			valueChangedChan <- v

		case vv := <-valueChangedChan:
			select {
			case value := <-vv:
				if err := value.GetError(); err != nil {
					time.Sleep(time.Second * 10)
					w.watchValueChan <- value.GetPath()
					break
				}
				content, _ := value.GetValue()
				w.cache.Set(value.GetPath(), content)
				w.watchValueChan <- value.GetPath()
			default:
			}

		}
	}
}
*/
