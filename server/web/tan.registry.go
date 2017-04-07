package web

func (s *WebServer) registerService() {
	if s.register != nil {
		addr := s.GetAddress()
		s.register.Register(addr, addr)

	}
}
func (s *WebServer) unRegisterService() {
	if s.register != nil {
		addr := s.GetAddress()
		s.register.UnRegister(addr, addr)

	}
}
