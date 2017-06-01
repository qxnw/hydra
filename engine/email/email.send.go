package email

import (
	"net/smtp"
	"strings"

	"github.com/qxnw/hydra/context"
)

func (s *emailProxy) sendMail(ctx *context.Context) (r string, t int, err error) {
	m, err := s.getGetParams(ctx)
	if err != nil {
		return
	}
	err = s.send(m)
	if err != nil {
		return
	}
	return "SUCCESS", 200, nil
}

func (s *emailProxy) send(m *email) error {
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
