package cron

import (
	"fmt"
	"sort"
	"strings"
	"time"

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
			LOOP:
				children, err = s.registry.WatchChildren(s.registryRoot)
				if err != nil {
					s.Errorf("监控服务节点发生错误:err:%v", err)
					if !s.running {
						return
					}
					time.Sleep(time.Second)
					goto LOOP
				}
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
	if strings.ToUpper(s.cluster) == "P2P" {
		return true
	}
	ncldrs := make([]string, 0, len(cldrs))
	for _, v := range cldrs {
		args := strings.SplitN(v, "_", 2)
		ncldrs = append(ncldrs, args[len(args)-1])
	}
	if len(ncldrs) == 0 {
		return false
	}
	sort.Strings(ncldrs)
	return strings.HasSuffix(s.clusterPath, ncldrs[0])
}
