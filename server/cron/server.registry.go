package cron

import "fmt"

func (s *CronServer) registryServer() (err error) {
	if s.registry == nil {
		return
	}
	s.clusterPath, err = s.registry.RegisterSeqNode(fmt.Sprintf("%s/%s_", s.registryRoot, s.ip), s.ip)

	if err != nil {
		return err
	}
	//	return nil
	return s.watchServerChange(s.clusterPath)
}
func (s *CronServer) watchServerChange(path string) error {
	cldrs, _, err := s.registry.GetChildren(s.registryRoot)
	if err != nil {
		return err
	}
	s.isMaster = len(cldrs) == 0 || (len(cldrs) > 0 && cldrs[0] == s.clusterName)
	s.Debug("cron.node.changed:", s.isMaster, cldrs)
	children, err := s.registry.WatchChildren(s.registryRoot)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-s.close:
				return
			case cldWatcher := <-children:
				cldrs, _ = cldWatcher.GetValue()
				s.isMaster = len(cldrs) == 0 || (len(cldrs) > 0 && cldrs[0] == s.clusterName)
				s.Debug("cron.node.changed:", s.isMaster, cldrs)
			}
		}
	}()
	return nil
}
func (s *CronServer) unregistryServer() {
	if s.registry != nil && s.clusterPath != "" {
		s.registry.Unregister(s.clusterPath)
	}
}
