package micro

import (
	"fmt"
	"os"
	"strings"

	"github.com/qxnw/lib4go/logger"
	"github.com/urfave/cli"

	_ "github.com/qxnw/hydra/client/rpc"
	_ "github.com/qxnw/hydra/engines"
	_ "github.com/qxnw/hydra/registry/zookeeper"
	_ "github.com/qxnw/hydra/servers/cron"
	_ "github.com/qxnw/hydra/servers/http"
	_ "github.com/qxnw/hydra/servers/mqc"
	_ "github.com/qxnw/hydra/servers/rpc"

	_ "github.com/qxnw/hydra/servers/http"
	_ "github.com/qxnw/lib4go/cache/memcache"
	_ "github.com/qxnw/lib4go/cache/redis"
	_ "github.com/qxnw/lib4go/mq/redis"
	_ "github.com/qxnw/lib4go/mq/stomp"
	_ "github.com/qxnw/lib4go/mq/xmq"
	_ "github.com/qxnw/lib4go/queue"
	_ "github.com/qxnw/lib4go/queue/redis"
)

var (
	microServerType   = []string{"api", "rpc", "web"}
	flowServerType    = []string{"cron", "mqc"}
	supportServerType = []string{"api", "rpc", "web", "cron", "mqc", "micro", "flow"}
)

//MicroApp  微服务应用
type MicroApp struct {
	app          *cli.App
	logger       *logger.Logger
	hydra        *Hydra
	platName     string
	systemName   string
	serverType   []string
	clusterName  string
	trace        string
	registryAddr string
	*option
}

//NewApp 创建微服务应用
func NewApp(opts ...Option) (m *MicroApp) {
	m = &MicroApp{app: cli.NewApp(), option: &option{}}
	for _, opt := range opts {
		opt(m.option)
	}
	m.logger = logger.GetSession("micro", logger.CreateSession())
	m.app.Name = "micro"
	m.app.Version = "2.0.0"
	m.app.Usage = "hydra微服务"
	cli.HelpFlag = cli.BoolFlag{
		Name:  "help,h",
		Usage: "查看帮助信息",
	}
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version,v",
		Usage: "查看micro版本信息",
	}
	m.app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: fmt.Sprintf("启动hydrd服务器,支持的服务器类型有：%s", supportServerType),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name,n",
					Usage: "服务器全称:格式:/平台名称/系统名称/服务器类型/集群名称",
				},
				cli.StringFlag{
					Name:  "registry,r",
					Usage: "注册中心地址:格式:proto://addr1,addr2",
				},
				cli.StringFlag{
					Name:        "trace,t",
					Destination: &m.trace,
					Usage:       "性能跟踪,可选参数:cpu,mem,block,mutex,web",
				}, cli.BoolFlag{
					Name: "debug,d",
					//Destination: &m.isDebug,
					Usage: "启用调试模式",
				},
			},
			Action: m.action,
		},
	}
	return m
}

//Start 启动服务器
func (m *MicroApp) Start() {
	defer logger.Close()
	if m.name != "" {
		if err := m.parsePath(m.name); err != nil {
			m.logger.Error(err)
			return
		}
	}
	if err := m.app.Run(os.Args); err != nil {
		return
	}
}

func (m *MicroApp) action(c *cli.Context) error {
	if err := m.checkInput(c.String("name"), c.String("registry")); err != nil {
		cli.ErrWriter.Write([]byte("  " + err.Error() + "\n\n"))
		cli.ShowCommandHelp(c, c.Command.Name)
		return nil
	}
	m.hydra = NewHydra(m.platName, m.systemName, m.serverType, m.clusterName, m.trace,
		m.registryAddr, m.isDebug)
	if err := m.hydra.Start(); err != nil {
		m.logger.Error(err)
		return err
	}
	return nil
}

func (m *MicroApp) checkInput(f string, registryAddr string) error {
	if f == "" && (m.platName == "" || m.systemName == "" || m.clusterName == "") {
		err := fmt.Errorf("--path不能为空")
		return err
	}
	if m.registryAddr == "" && registryAddr == "" {
		err := fmt.Errorf("--registry不能为空")
		return err
	}
	if registryAddr != "" {
		m.registryAddr = registryAddr
	}
	return m.parsePath(f)

}

func (m *MicroApp) parsePath(p string) (err error) {
	fs := strings.Split(strings.Trim(p, "/"), "/")
	if len(fs) != 4 {
		err := fmt.Errorf("系统名称错误，格式:/[platName]/[sysName]/[typeName]/[clusterName]")
		return err
	}
	if m.serverType, err = m.getServerTypes(strings.Split(fs[2], "-")); err != nil {
		return err
	}
	m.platName = fs[0]
	m.systemName = fs[1]
	m.clusterName = fs[3]
	return nil
}
func (m *MicroApp) getServerTypes(sts []string) ([]string, error) {
	removeRepMap := make(map[string]byte)
	for _, v := range sts {
		var ctn bool
		for _, k := range supportServerType {
			if ctn = k == v; ctn {
				break
			}
		}
		if !ctn {
			return nil, fmt.Errorf("不支持的服务器类型:%v", v)
		}
		switch v {
		case "*":
			for _, value := range microServerType {
				removeRepMap[value] = 0
			}
			for _, value := range flowServerType {
				removeRepMap[value] = 0
			}
			break
		case "micro":
			for _, value := range microServerType {
				removeRepMap[value] = 0
			}
		case "flow":
			for _, value := range flowServerType {
				removeRepMap[value] = 0
			}
		default:
			removeRepMap[v] = 0
		}
	}
	types := make([]string, 0, len(removeRepMap))
	for k := range removeRepMap {
		types = append(types, k)
	}
	return types, nil
}
