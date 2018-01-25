package ssm

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func loadService(r *component.StandardComponent, i component.IContainer) {
	r.AddMicroService("/ssm/ytx/send", SendYTXSMS(), "ssm")
	r.AddMicroService("/ssm/wx/send", SendWeiXinMesssage(i), "ssm")
	r.AddMicroService("/ssm/email/send", SendMail(i), "ssm")

}
func init() {
	engines.AddServiceLoader("ssm", loadService)
}
