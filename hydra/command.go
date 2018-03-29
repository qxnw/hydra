package hydra

import (
	"fmt"

	"github.com/urfave/cli"
)

//VERSION 版本号
var VERSION string = "2.0.0"

func (m *MicroApp) getCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "hydra"
	app.Version = VERSION
	app.Usage = "hydra微服务"
	cli.HelpFlag = cli.BoolFlag{
		Name:  "help,h",
		Usage: "查看帮助信息",
	}
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version,v",
		Usage: "查看hydra版本信息",
	}
	app.Commands = m.getCommands()
	return app
}

func (m *MicroApp) getCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "start",
			Usage:  "启动hydra服务器",
			Flags:  m.getStartFlags(),
			Action: m.action,
		},
	}
}
func (m *MicroApp) getStartFlags() []cli.Flag {
	flags := make([]cli.Flag, 0, 4)
	if m.RegistryAddr == "" {
		flags = append(flags, cli.StringFlag{
			Name:        "registry,r",
			Destination: &m.RegistryAddr,
			Usage:       "注册中心:格式:proto://addr1,addr2",
		})
	}
	if m.PlatName == "" && m.SystemName == "" && len(m.ServerTypes) == 0 && m.ClusterName == "" {
		flags = append(flags, cli.StringFlag{
			Name:  "name,n",
			Usage: "服务全称:格式:/平台名称/系统名称/服务器类型/集群名称",
		})
	} else {
		if m.PlatName == "" {
			flags = append(flags, cli.StringFlag{
				Name:        "plat,p",
				Destination: &m.PlatName,
				Usage:       "平台名称",
			})
		}
		if m.SystemName == "" {
			flags = append(flags, cli.StringFlag{
				Name:        "system,s",
				Destination: &m.SystemName,
				Usage:       "系统名称",
			})
		}
		if len(m.ServerTypes) == 0 {
			flags = append(flags, cli.StringFlag{
				Name:        "serverType,t",
				Destination: &m.ServerTypeNames,
				Usage:       fmt.Sprintf("服务类型%v", supportServerType),
			})
		}
		if m.ClusterName == "" {
			flags = append(flags, cli.StringFlag{
				Name:        "cluster,c",
				Destination: &m.ClusterName,
				Usage:       "集群名称",
			})
		}
	}

	flags = append(flags, cli.StringFlag{
		Name:        "trace",
		Destination: &m.Trace,
		Usage:       fmt.Sprintf("性能跟踪%v", supportTraces),
	})
	flags = append(flags, cli.BoolFlag{
		Name:        "remote-logger,l",
		Destination: &m.remoteLogger,
		Usage:       "启用远程日志",
	})

	return flags
}
