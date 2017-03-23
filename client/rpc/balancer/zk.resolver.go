package balancer

import (
	"fmt"
	"strings"
	"time"

	"github.com/qxnw/lib4go/zk"

	"google.golang.org/grpc/naming"
)

//ZKResolver 服务解析器
type ZKResolver struct {
	timeout time.Duration
	service string
	local   string
}

//NewZKResolver 返回服务解析器
func NewZKResolver(service string, local string, timeout time.Duration) naming.Resolver {
	return &ZKResolver{timeout: timeout, service: service, local: local}
}

// Resolve to resolve the service from zookeeper, target is the dial address of zookeeper
// target example: "192.168.0.159:2181;192.168.0.154:2181"
func (re *ZKResolver) Resolve(target string) (naming.Watcher, error) {
	client, err := zk.New(strings.Split(target, ";"), re.timeout)
	if err != nil {
		return nil, fmt.Errorf("grpclb: create zookeeper client failed: %s", err.Error())
	}
	return &ZKWatcher{re: re, client: client, local: re.local}, nil
}
