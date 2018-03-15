package api

import (
	"os"

	"github.com/qxnw/lib4go/logger"
	"github.com/urfave/cli"
)

//MicroApp  微服务
type MicroApp struct {
	app      *cli.App
	logger   *logger.Logger
	platName string
}

//NewMicroApp 创建微服务应用
func NewMicroApp() (m *MicroApp) {
	m = &MicroApp{app: cli.NewApp()}
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
					Name: "plat",
					//Aliases: []string{"plat", "p"},
					Usage: "服务器配置路径",
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
	return nil
}
