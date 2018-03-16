package conf

import "sync"

type metadata struct {
	data map[string]interface{}
	lock sync.RWMutex
}

func (m *metadata) Get(key string) interface{} {
	m.lock.RLock()
	data := m.data[key]
	m.lock.RUnlock()
	return &data
}
func (m *metadata) Set(key string, value interface{}) {
	m.lock.Lock()
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
	m.lock.Unlock()
}

type MetadataConf struct {
	Name     string
	Type     string
	metadata metadata
}

func (s *MetadataConf) GetMetadata(key string) interface{} {
	return s.metadata.Get(key)
}
func (s *MetadataConf) SetMetadata(key string, v interface{}) {
	s.metadata.Set(key, v)
}
