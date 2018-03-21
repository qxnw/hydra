package hydra

import (
	"fmt"
	"os"
	"reflect"

	"github.com/asaskevich/govalidator"
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

//MicroApp  微服务应用
type MicroApp struct {
	app    *cli.App
	logger *logger.Logger
	hydra  *Hydra
	*option
}

//NewApp 创建微服务应用
func NewApp(opts ...Option) (m *MicroApp) {
	m = &MicroApp{option: &option{}}
	for _, opt := range opts {
		opt(m.option)
	}
	m.logger = logger.GetSession("micro", logger.CreateSession())

	return m
}

//Start 启动服务器
func (m *MicroApp) Start() {
	defer logger.Close()
	m.app = m.getCliApp()
	if err := m.app.Run(os.Args); err != nil {
		return
	}
}

func (m *MicroApp) action(c *cli.Context) error {
	if err := m.checkInput(); err != nil {
		cli.ErrWriter.Write([]byte("  " + err.Error() + "\n\n"))
		cli.ShowCommandHelp(c, c.Command.Name)
		return nil
	}
	m.hydra = NewHydra(m.PlatName, m.SystemName, m.ServerTypes, m.ClusterName, m.Trace,
		m.RegistryAddr, m.IsDebug)
	if err := m.hydra.Start(); err != nil {
		m.logger.Error(err)
		return err
	}
	return nil
}

func (m *MicroApp) checkInput() (err error) {
	if b, err := govalidator.ValidateStruct(m.option); !b {
		err = fmt.Errorf("validate(%v) %v", reflect.TypeOf(m.option), err)
		return err
	}
	return
}
