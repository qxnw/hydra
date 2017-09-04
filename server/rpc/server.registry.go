package rpc

import (
	"fmt"
	"strings"
)

func (s *RPCServer) registerService() (err error) {
	if s.registry == nil {
		return
	}
	s.localRRCServices = s.getServiceList(s.localRRCServices)
	for _, v := range s.localRRCServices {
		path, err := s.registry.RegisterService(v, strings.Replace(s.GetAddress(), "//", "", -1), s.GetAddress())
		if err != nil {
			return err
		}
		s.remoteRPCService = append(s.remoteRPCService, path)
	}
	addr := s.GetAddress()
	s.clusterPath, err = s.registry.RegisterSeqNode(fmt.Sprintf("%s/%s_", s.registryRoot, s.GetAddress()), addr)
	return nil
}
func (s *RPCServer) unRegisterService() {
	if s.registry == nil {
		return
	}
	if s.clusterPath != "" {
		s.registry.Unregister(s.clusterPath)
	}
	for _, v := range s.remoteRPCService {
		s.registry.Unregister(v)
	}
	s.remoteRPCService = make([]string, 0, 8)
}
