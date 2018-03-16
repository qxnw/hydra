package cron

import (
	"sort"
	"strings"
	"time"

	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/lib4go/types"
)

func (s *CronResponsiveServer) watchMasterChange(root, path string) error {
	cldrs, _, err := s.engine.GetRegistry().GetChildren(root)
	if err != nil {
		return err
	}
	s.master = s.isMaster(path, cldrs)
	servers.Tracef(s.Debugf, "%s是:%s", s.currentConf.GetServerName(), types.DecodeString(s.master, true, "master server", "slave server"))
	if err = s.notifyConsumer(s.master); err != nil {
		return err
	}
	children, err := s.engine.GetRegistry().WatchChildren(root)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-s.closeChan:
				return
			case cldWatcher := <-children:
				cldrs, _ = cldWatcher.GetValue()
				master := s.isMaster(path, cldrs)
				if master != s.master {
					servers.Tracef(s.Debugf, "%s是:%s", s.currentConf.GetServerName(), types.DecodeString(master, true, "master server", "slave server"))
					s.notifyConsumer(master)
					s.master = master
				}

			LOOP:
				children, err = s.engine.GetRegistry().WatchChildren(root)
				if err != nil {
					servers.Tracef(s.Errorf, "%s:监控服务节点发生错误:err:%v", s.currentConf.GetServerName(), err)
					if s.done {
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

func (s *CronResponsiveServer) isMaster(path string, cldrs []string) bool {
	ncldrs := make([]string, 0, len(cldrs))
	for _, v := range cldrs {
		args := strings.SplitN(v, "_", 2)
		ncldrs = append(ncldrs, args[len(args)-1])
	}
	sort.Strings(ncldrs)
	s.shardingCount = s.currentConf.GetInt("sharding", 0)
	if s.shardingCount == 0 {
		s.shardingCount = len(ncldrs)
	}
	index := -1
	for i, v := range ncldrs {
		if strings.HasSuffix(path, v) {
			index = i
			break
		}
	}
	s.shardingIndex = getSharding(index, s.shardingCount)
	return s.shardingIndex > -1

}
func (s *CronResponsiveServer) notifyConsumer(v bool) error {
	if v {
		return s.server.Resume()
	}
	s.server.Pause()
	return nil
}
func getSharding(index int, count int) int {
	if count <= 0 && index >= 0 {
		return index
	}
	if index < 0 || index >= count {
		return -1
	}
	return index % count
}
