package context

import (
	"fmt"

	"github.com/qxnw/lib4go/utility"
)

type IData interface {
	Set(string, string)
	Get(string) (string, error)
	Each(func(string, string))
}

//Request 输入参数
type Request struct {
	Form           *inputParams
	Param          *inputParams
	Setting        *inputParams
	CircuitBreaker *circuitBreakerParam //熔断处理
	Http           *httpRequest
	Ext            *extParams
}

func newRequest() *Request {
	return &Request{
		Form:           &inputParams{},
		Param:          &inputParams{},
		Setting:        &inputParams{},
		CircuitBreaker: &circuitBreakerParam{inputParams: &inputParams{}},
		Http:           &httpRequest{},
		Ext:            &extParams{},
	}
}

func (r *Request) reset(form IData, param IData, setting IData, ext map[string]interface{}) {
	r.Form.data = form
	r.Param.data = param
	r.Setting.data = param
	r.CircuitBreaker.inputParams.data = setting
	r.CircuitBreaker.ext = ext
	r.Ext.ext = ext
	r.Http.ext = ext

}

//Check 检查输入参数和配置参数是否为空
func (w *Request) Check(checker map[string][]string) (int, error) {
	if err := w.Form.Check(checker["input"]...); err != nil {
		return ERR_NOT_ACCEPTABLE, fmt.Errorf("输入参数:%v", err)
	}
	if err := w.Setting.Check(checker["setting"]...); err != nil {
		return ERR_NOT_EXTENDED, fmt.Errorf("配置参数:%v", err)
	}
	return 0, nil
}

//Body2Input 根据编码格式解码body参数，并更新input参数
func (w *Request) Body2Input(encoding ...string) error {
	body, err := w.Ext.GetBody(encoding...)
	if err != nil {
		return err
	}
	qString, err := utility.GetMapWithQuery(body)
	if err != nil {
		return err
	}
	for k, v := range qString {
		w.Form.data.Set(k, v)
	}
	return nil
}

//clear 清空数据
func (r *Request) clear() {
	r.Form.data = nil
	r.Param.data = nil
	r.Setting.data = nil
	r.CircuitBreaker.inputParams = nil
	r.CircuitBreaker.ext = nil
	r.Ext.ext = nil
	r.Http.ext = nil
}
