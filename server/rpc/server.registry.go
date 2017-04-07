package rpc

func (s *RPCServer) registerService() {
	if s.registry != nil {
		addr := s.GetAddress()
		for _, v := range s.services {
			s.registry.Register(v, addr)
		}
	}
}
func (s *RPCServer) unRegisterService() {
	if s.registry != nil {
		addr := s.GetAddress()
		for _, v := range s.services {
			s.registry.UnRegister(v, addr)
		}
	}
}
