package email

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
)

type emailProxy struct {
	domain     string
	serverName string
	serverType string
	services   []string
}

func newEmailProxy() *emailProxy {
	return &emailProxy{
		services: []string{"/email/send"},
	}
}

func (s *emailProxy) Start(ctx *engine.EngineContext) (services []string, err error) {
	s.domain = ctx.Domain
	s.serverName = ctx.ServerName
	s.serverType = ctx.ServerType
	return s.services, nil
}
func (s *emailProxy) Close() error {
	return nil
}

//Handle
//从input参数中获取 receiver,subject,content
//从args参数中获取 mail
//配置文件格式:{"smtp":"smtp.exmail.qq.com:25", "sender":"yanglei@100bm.cn","password":"12333"}

func (s *emailProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	m, err := s.getGetParams(ctx)
	if err != nil {
		err = fmt.Errorf("engine:email.%v", err)
		return
	}
	err = s.sendMail(m)
	if err != nil {
		return
	}
	r = &context.Response{Status: 200, Content: "SUCCESS"}
	return
}

func (s *emailProxy) sendMail(m *email) error {
	auth := smtp.PlainAuth("", m.sender, m.password, m.host)
	msg := []byte("To: " + strings.Join(m.receiver, ";") + "\r\nFrom: " +
		m.sender + "<" + m.sender + ">\r\nSubject: " + m.subject + "\r\n" + m.mailtype + "\r\n\r\n" + m.content)

	err := smtp.SendMail(
		m.smtp,
		auth,
		m.sender,
		m.receiver,
		msg,
	)
	return err
}
func (s *emailProxy) Has(shortName, fullName string) (err error) {
	for _, v := range s.services {
		if v == shortName {
			return nil
		}
	}
	return fmt.Errorf("不存在服务:%s", shortName)
}

type emailProxyResolver struct {
}

func (s *emailProxyResolver) Resolve() engine.IWorker {
	return newEmailProxy()
}

func init() {
	engine.Register("email", &emailProxyResolver{})
}
