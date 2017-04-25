package mq

func (s *MQConsumer) registryServer() (err error) {
	if s.registry != nil {
		s.clusterPath, err = s.registry.RegisterWithPath(s.registryRoot, s.ip)
		return
	}
	return
}
func (s *MQConsumer) unregistryServer() {
	if s.registry != nil && s.clusterPath != "" {
		s.registry.Unregister(s.clusterPath)
	}
}
