package api

func (s *HTTPServer) registryServer() (err error) {
	if s.registry != nil {
		addr := s.GetAddress()
		s.clusterPath, err = s.registry.RegisterSeqNode(s.registryRoot, addr)
		return
	}
	return
}
func (s *HTTPServer) unregistryServer() {
	if s.registry != nil && s.clusterPath != "" {
		s.registry.Unregister(s.clusterPath)
	}
}
