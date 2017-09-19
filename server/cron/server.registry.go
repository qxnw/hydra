package cron

import (
	"fmt"
	"sort"
	"strings"

	"github.com/qxnw/lib4go/types"
)

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
	s.isMaster = s.isMasterCron(cldrs)
	s.Debugf("当前%s(cron)是%s", s.serverName, types.DecodeString(s.isMaster, true, "master", "slave"))
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
				master := s.isMasterCron(cldrs)
				if master != s.isMaster {
					s.Debugf("当前%s(cron)是%s", s.serverName, types.DecodeString(master, true, "master", "slave"))
				}
				s.isMaster = master
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
func (s *CronServer) isMasterCron(cldrs []string) bool {
	sort.Strings(cldrs)
	return strings.HasSuffix(s.clusterPath, cldrs[0])
}
