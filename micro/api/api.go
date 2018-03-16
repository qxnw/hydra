package api

import (
	"os"

	"github.com/qxnw/hydra/micro"
	"github.com/qxnw/lib4go/logger"
	"github.com/urfave/cli"

	_ "github.com/qxnw/hydra/servers/http"
	_ "github.com/qxnw/lib4go/cache/memcache"
	_ "github.com/qxnw/lib4go/cache/redis"
	_ "github.com/qxnw/lib4go/mq/redis"
	_ "github.com/qxnw/lib4go/mq/stomp"
	_ "github.com/qxnw/lib4go/mq/xmq"
	_ "github.com/qxnw/lib4go/queue"
	_ "github.com/qxnw/lib4go/queue/redis"
)

//MicroApp  微服务
type MicroApp struct {
	app          *cli.App
	logger       *logger.Logger
	hydra        *micro.Hydra
	platName     string
	systemName   string
	serverType   string
	clusterName  string
	trace        string
	registryAddr string
	isDebug      bool
}

//NewMicroApp 创建微服务应用
func NewMicroApp(platName string, systemName string) (m *MicroApp) {
	m = &MicroApp{app: cli.NewApp(), serverType: "api", platName: platName, systemName: systemName}
	m.app.Name = "micro"
	m.app.Usage = "启动微服务"
	m.logger = logger.GetSession("micro", logger.CreateSession())
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version,v",
		Usage: "查看micro版本信息",
	}
	m.app.Commands = []cli.Command{
		{
			Name:    "start",
			Aliases: []string{"start"},
			Usage:   "启动服务器",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "t",
					Value:       "*",
					Destination: &m.clusterName,
					Usage:       "集群名称，默认为'*'",
				}, cli.StringFlag{
					Name:        "p",
					Value:       "",
					Destination: &m.trace,
					Usage:       "性能跟踪",
				}, cli.BoolFlag{
					Name:        "d",
					Destination: &m.isDebug,
					Usage:       "是否启用调试",
				},
			},
			Action: m.action,
		},
	}
	return m
}

//Start 启动服务器
func (m *MicroApp) Start() {
	if err := m.app.Run(os.Args); err != nil {
		m.logger.Error(err)
	}
}

func (m *MicroApp) action(c *cli.Context) error {
	m.hydra = micro.NewHydra(m.platName, m.systemName, []string{m.serverType}, m.clusterName, m.trace,
		m.registryAddr, m.isDebug)
	return m.hydra.Start()
}
