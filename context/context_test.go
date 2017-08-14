package context

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	//_ "github.com/mattn/go-oci8"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/ut"
	"github.com/qxnw/lib4go/utility"
)

//Context 服务输出及Task执行的上下文
type context struct {
	Input inputArgs
	Ext   map[string]interface{}
}

//InputArgs 上下文输入参数
type inputArgs struct {
	Input  transform.ITransformGetter "input"`
	Body   string                     `json:"body"`
	Params transform.ITransformGetter `json:"params"`
	Args   map[string]string          `json:"args"`
}

func (c *inputArgs) ToJson() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func (c *context) GetInput() transform.ITransformGetter {
	return c.Input.Input
}
func (c *context) GetArgs() map[string]string {
	return c.Input.Args
}
func (c *context) GetBody() string {
	return c.Input.Body
}
func (c *context) GetParams() transform.ITransformGetter {
	return c.Input.Params
}
func (c *context) GetJson() string {
	return c.Input.ToJson()
}
func (c *context) GetExt() map[string]interface{} {
	return c.Ext
}
func newContext() (r *context) {
	r = &context{}
	r.Input.Input = transform.New().Data
	r.Input.Params = r.Input.Input
	r.Input.Body = ""
	r.Input.Args = map[string]string{
		"db":    "oracle",
		"cache": "mem",
	}

	r.Ext = map[string]interface{}{
		"hydra_sid": utility.GetGUID(),
		"__func_var_get_": func(c string, n string) (string, error) {
			fmt.Println(c, n)
			if n == "oracle" {
				return `{
    "provider":"oracle",
    "connString":"wx_base_system/123456@orcl136"
}`, nil
			} else if n == "mem" {
				return `{"server":"192.168.0.166:11212"}`, nil
			}
			return "", errors.New("未找到--------")
		},
	}
	return
}
func TestContextDb(t *testing.T) {
	context, err := GetContext(newContext(), nil)
	ut.ExpectSkip(t, err, nil)
	db, err := context.GetDB()
	ut.ExpectSkip(t, err, nil)
	ut.RefuteSkip(t, db, nil)
	s, _, _, err := db.Scalar("select 1 from dual", make(map[string]interface{}))
	ut.ExpectSkip(t, err, nil)
	ut.ExpectSkip(t, fmt.Sprintf("%v", s), "1")
}

func TestContextCache(t *testing.T) {
	context, err := GetContext(newContext(), nil)
	ut.ExpectSkip(t, err, nil)
	db, err := context.GetCache()
	ut.ExpectSkip(t, err, nil)
	_, err = db.Get("abc")
	ut.RefuteSkip(t, err, nil)
}
