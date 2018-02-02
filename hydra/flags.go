package hydra

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/lib4go/net"
	"github.com/qxnw/lib4go/transform"
	"github.com/spf13/pflag"
)

//HFlags hydra输入参数绑定
type HFlags struct {
	currentRegistry        string
	Domain                 string
	runMode                string
	tag                    string
	mask                   string
	trace                  string
	crossRegistry          string
	rpcLogger              bool
	currentRegistryAddress []string
	crossRegistryAddress   []string
	ip                     string
	baseData               *transform.Transform
	flag                   *pflag.FlagSet
	inputArgs              []string
}

//BindFlags 绑定参数列表
func (h *HFlags) BindFlags(flag *pflag.FlagSet) {
	flag.StringVarP(&h.currentRegistry, "registry center address", "r", "", "注册中心地址(格式：zk://192.168.0.159:2181,192.168.0.158:2181)")
	flag.StringVarP(&h.mask, "ip mask", "i", "", "ip掩码(本有多个IP时指定，格式:192.168.0)")
	flag.StringVarP(&h.tag, "server tag", "t", "", "服务器名称(默认为本机IP地址)")
	flag.StringVarP(&h.trace, "enable trace", "p", "", "启用项目性能跟踪cpu/mem/block/mutex/server")
	flag.BoolVarP(&servers.IsDebug, "enable debug", "d", false, "是否启用调试模式")
	flag.StringVarP(&h.crossRegistry, "cross  registry  center address", "c", "", "跨域注册中心地址")
	flag.BoolVarP(&h.rpcLogger, "use rpc logger", "g", false, "使用RPC远程记录日志")
	h.flag = flag

}

//CheckFlags 检查输入参数
func (h *HFlags) CheckFlags(i ...int) (err error) {
	h.inputArgs = make([]string, 0, len(os.Args))
	for _, v := range os.Args {
		h.inputArgs = append(h.inputArgs, v)
	}
	h.flag.Parse(os.Args[1:])
	index := 1
	if len(i) > 0 {
		index = i[0]
	}
	if len(os.Args) <= index {
		return errors.New("未指定域名称")
	}
	h.Domain = os.Args[index]
	ds := strings.Split(strings.Trim(h.Domain, "/"), "/")
	if len(ds) == 1 {
		h.Domain = ds[0]
	} else if len(ds) == 2 {
		h.Domain = ds[0]
	} else {
		err = fmt.Errorf("域名称配置错误:%s", os.Args[index])
		return
	}

	if h.currentRegistry == "" {
		h.runMode = modeStandalone
		h.currentRegistryAddress = []string{"localhost"}
		h.currentRegistry = fmt.Sprintf("%s://%s", h.runMode, strings.Join(h.currentRegistryAddress, ","))
	} else {
		h.runMode = modeCluster
		h.runMode, h.currentRegistryAddress, err = registry.ResolveAddress(h.currentRegistry)
		if err != nil {
			return fmt.Errorf("集群地址配置有误:%v", err)
		}
	}
	if h.crossRegistry != "" {
		if strings.Contains(h.crossRegistry, "//") {
			return fmt.Errorf("跨域注册中心地址不能指定协议信息:%s(err:%v)", h.crossRegistry, err)
		}
		h.crossRegistryAddress = strings.Split(h.crossRegistry, ",")
	}
	h.ip = net.GetLocalIPAddress(h.mask)
	if h.tag == "" {
		h.tag = h.ip
	}
	h.baseData = transform.NewMap(map[string]string{
		"ip": h.ip,
	})
	h.tag = h.baseData.Translate(h.tag)
	return nil
}

//ToArgs 转换为可执行的参数列表
func (h *HFlags) ToArgs() []string {
	args := make([]string, 0, 4)
	args = append(args, h.Domain)
	args = append(args, "-r", h.currentRegistry)
	if h.tag != "" {
		args = append(args, "-t", h.tag)
	}
	if h.trace != "" {
		args = append(args, "-p", h.trace)
	}
	if h.mask != "" {
		args = append(args, "-i", h.mask)
	}
	if servers.IsDebug {
		args = append(args, "-d")
	}
	if h.crossRegistry != "" {
		args = append(args, "-c", h.crossRegistry)
	}
	if h.rpcLogger {
		args = append(args, "-g")
	}
	return args
}
