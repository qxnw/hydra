package responsive

//IsValueChanged 检查值是否发生变化
func (s *ResponsiveConf) IsValueChanged(names ...string) (isChanged bool) {
	for _, name := range names {
		if s.Nconf.String(name) != s.Oconf.String(name) {
			return true
		}
	}
	return false
}

//HasNode 检查节点是否存在
func (s *ResponsiveConf) HasNode(name ...string) bool {
	for _, v := range name {
		if s.Nconf.Has("#@path/" + v) {
			return true
		}
	}
	return false
}

//IsNodeChanged 检查节点是否发生变化
func (s *ResponsiveConf) IsNodeChanged(name string) (isChanged bool) {
	oldExists := s.Oconf.Has("#@path/" + name)
	newExists := s.Nconf.Has("#@path/" + name)
	if oldExists != newExists {
		return true
	}
	if !newExists {
		return false
	}

	oldNode, err0 := s.Oconf.GetNodeWithSectionName(name, "#@path/"+name)
	newNode, err1 := s.Nconf.GetNodeWithSectionName(name, "#@path/"+name)
	if err1 != nil || err0 != nil {
		return true
	}
	//检查头配置是否变化，已变化则需要重启服务
	if newNode.GetVersion() != oldNode.GetVersion() {
		return true
	}
	return false
}

//IsRequiredNodeChanged 检查必须节点是否发生变化
func (s *ResponsiveConf) IsRequiredNodeChanged(name string) (isChanged bool, err error) {
	newNode, err1 := s.Nconf.GetNodeWithSectionName(name, "#@path/"+name)
	if err1 != nil {
		return true, err1
	}
	if s.Oconf.Has("#@path/"+name) != s.Nconf.Has("#@path/"+name) {
		return true, nil
	}
	oldNode, _ := s.Oconf.GetNodeWithSectionName(name, "#@path/"+name)
	//检查头配置是否变化，已变化则需要重启服务
	if newNode.GetVersion() != oldNode.GetVersion() {
		return true, nil
	}
	return false, nil
}
