package email

import (
	"errors"
	"fmt"

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

func (s *emailProxy) getGetParams(ctx *context.Context) (mail *email, err error) {
	mail = &email{mailtype: "Content-Type: text/plain; charset=UTF-8"}
	receivers, err := ctx.GetInput().Get("receiver")
	if err != nil || receivers == "" {
		err = fmt.Errorf("邮件接收人不能为空")
		return
	}
	mail.receiver = strings.Split(receivers, ";")
	mail.subject, err = ctx.GetInput().Get("subject")
	if err != nil || mail.subject == "" {
		err = fmt.Errorf("邮件标题不能为空")
		return
	}
	mail.content, err = ctx.GetInput().Get("content")
	if err != nil || mail.content == "" {
		err = fmt.Errorf("邮件内容不能为空")
		return
	}
	setting, ok := ctx.GetArgs()["setting"]
	if !ok {
		err = fmt.Errorf("邮件Args.setting配置不能为空")
		return
	}
	content, err := s.getVarParam(ctx, setting)
	if err != nil {
		return
	}

	settingData, err := jsons.Unmarshal([]byte(content))
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，无法解析(err:%v)", setting, err)
		return
	}
	if settingData["smtp"] == nil ||
		settingData["sender"] == nil || settingData["password"] == nil {
		err = fmt.Errorf("setting[%s]配置错误，未配置字段:host,smtp,sender,password", setting)
		return
	}

	mail.smtp, ok = settingData["smtp"].(string)
	if !ok {
		err = fmt.Errorf("setting[%s]配置错误，未配置字段:smtp", setting)
		return
	}
	mail.host = strings.SplitN(mail.smtp, ":", 2)[0]
	mail.sender, ok = settingData["sender"].(string)
	if !ok {
		err = fmt.Errorf("setting[%s]配置错误，未配置字段:sender", setting)
		return
	}
	mail.password, ok = settingData["password"].(string)
	if !ok {
		err = fmt.Errorf("setting[%s]配置错误，未配置字段:password", setting)
		return
	}
	return

}

func (s *emailProxy) getVarParam(ctx *context.Context, name string) (string, error) {
	funcVar := ctx.GetExt()["__func_var_get_"]
	if funcVar == nil {
		return "", errors.New("未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f("setting", name)
	}
	return "", errors.New("未找到__func_var_get_传入类型错误")
}
