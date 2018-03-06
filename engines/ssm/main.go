package ssm

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func loadService(r *component.StandardComponent, c component.IContainer) {
	r.AddMicroService("/ssm/ytx/send", SendYTXSMS(c), "ssm")
	r.AddMicroService("/ssm/wx/send", SendWeiXinMesssage(c), "ssm")
	r.AddMicroService("/ssm/email/send", SendMail(c), "ssm")

}
func init() {
	engines.AddLoader("ssm", loadService)
}
