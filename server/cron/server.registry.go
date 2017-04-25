package cron

func (s *CronServer) registryServer() (err error) {
	if s.registry != nil {
		s.clusterPath, err = s.registry.RegisterWithPath(s.registryRoot, s.ip)
		return
	}
	return
}
func (s *CronServer) unregistryServer() {
	if s.registry != nil && s.clusterPath != "" {
		s.registry.Unregister(s.clusterPath)
	}
}
