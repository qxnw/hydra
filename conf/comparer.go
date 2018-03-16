package conf

type Comparer struct {
	Oconf IServerConf
	Nconf IServerConf
}

func NewComparer(Oconf IServerConf, Nconf IServerConf) *Comparer {
	return &Comparer{
		Oconf: Oconf,
		Nconf: Nconf,
	}
}
func (s *Comparer) IsChanged() bool {
	return s.Oconf.GetVersion() != s.Nconf.GetVersion()
}

//IsValueChanged 检查值是否发生变化
func (s *Comparer) IsValueChanged(names ...string) (isChanged bool) {
	for _, name := range names {
		if s.Nconf.GetString(name) != s.Oconf.GetString(name) {
			return true
		}
	}
	return false
}

//IsSubConfChanged 检查节点是否发生变化
func (s *Comparer) IsSubConfChanged(name string) (isChanged bool) {
	oldConf, _ := s.Oconf.GetSubConf(name)
	newConf, _ := s.Nconf.GetSubConf(name)
	if oldConf == nil {
		oldConf = &JSONConf{version: 0}
	}
	if newConf == nil {
		newConf = &JSONConf{version: 0}
	}
	return oldConf.version != newConf.version
}

//IsRequiredSubConfChanged 检查必须节点是否发生变化
func (s *Comparer) IsRequiredSubConfChanged(name string) (isChanged bool, err error) {
	oldConf, _ := s.Oconf.GetSubConf(name)
	newConf, err := s.Nconf.GetSubConf(name)
	if err != nil {
		return true, err
	}
	if oldConf == nil {
		oldConf = &JSONConf{version: 0}
	}
	if newConf == nil {
		newConf = &JSONConf{version: 0}
	}
	return oldConf.GetVersion() != newConf.GetVersion(), nil
}
