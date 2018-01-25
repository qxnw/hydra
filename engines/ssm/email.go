package ssm

import (
	"fmt"
	"net/smtp"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"

	"strings"

	"github.com/qxnw/lib4go/jsons"
)

type email struct {
	receiver []string
	subject  string
	content  string
	mailtype string
	sender   string
	password string
	host     string
	smtp     string
}

func getEmailParams(ctx *context.Context) (mail *email, err error) {
	mail = &email{mailtype: "Content-Type: text/plain; charset=UTF-8"}
	receivers, err := ctx.Request.Form.Get("receiver")
	if err != nil || receivers == "" {
		err = fmt.Errorf("邮件接收人不能为空")
		return
	}
	mail.receiver = strings.Split(receivers, ";")
	mail.subject, err = ctx.Request.Form.Get("subject")
	if err != nil || mail.subject == "" {
		err = fmt.Errorf("邮件标题不能为空")
		return
	}
	mail.content, err = ctx.Request.Form.Get("content")
	if err != nil || mail.content == "" {
		err = fmt.Errorf("邮件内容不能为空")
		return
	}
	content, err := ctx.Request.Ext.GetVarParam("ssm", ctx.Request.Setting.GetString("email"))
	if err != nil {
		return
	}

	settingData, err := jsons.Unmarshal([]byte(content))
	if err != nil {
		err = fmt.Errorf("args.setting配置错误，无法解析(err:%v)", err)
		return
	}
	if settingData["smtp"] == nil ||
		settingData["sender"] == nil || settingData["password"] == nil {
		err = fmt.Errorf("args.setting配置错误，未配置字段:host,smtp,sender,password(%s)", content)
		return
	}
	var ok bool
	mail.smtp, ok = settingData["smtp"].(string)
	if !ok {
		err = fmt.Errorf("args.setting配置错误，未配置字段:smtp(%s)", content)
		return
	}
	mail.host = strings.SplitN(mail.smtp, ":", 2)[0]
	mail.sender, ok = settingData["sender"].(string)
	if !ok {
		err = fmt.Errorf("args.setting配置错误，未配置字段:sender(%s)", content)
		return
	}
	mail.password, ok = settingData["password"].(string)
	if !ok {
		err = fmt.Errorf("args.setting配置错误，未配置字段:password(%s)", content)
		return
	}
	return

}

func send(m *email) error {
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

//SendMail 发送邮件
func SendMail(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		m, err := getEmailParams(ctx)
		if err != nil {
			return
		}
		err = send(m)
		if err != nil {
			return
		}
		response.Success()
		return
	}
}
