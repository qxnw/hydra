package middleware

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/lib4go/encoding"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/utility"
)

func getUUID(c *gin.Context) string {
	ck, err := c.Request.Cookie("hydra_sid")
	if err != nil || ck == nil || ck.Value == "" {
		return logger.CreateSession()
	}
	return ck.Value
}
func setStartTime(c *gin.Context) {
	c.Set("__start_time_", time.Now())
}
func setLogger(c *gin.Context, l *logger.Logger) {
	c.Set("__logger_", l)
}
func getLogger(c *gin.Context) *logger.Logger {
	l, _ := c.Get("__logger_")
	return l.(*logger.Logger)
}
func getExpendTime(c *gin.Context) time.Duration {
	start, _ := c.Get("__start_time_")
	return time.Since(start.(time.Time))

}
func getJWTRaw(c *gin.Context) interface{} {
	jwt, _ := c.Get("__jwt_")
	return jwt
}
func setJWTRaw(c *gin.Context, v interface{}) {
	c.Set("__jwt_", v)
}
func getServiceName(c *gin.Context) string {
	if service, ok := c.Get("__service_"); ok {
		return service.(string)
	}
	return ""
}
func setServiceName(c *gin.Context, v string) {
	c.Set("__service_", v)
}
func setResponse(c *gin.Context, r context.Response) {
	c.Set("__response_", r)
}
func getResponse(c *gin.Context) context.Response {
	result, _ := c.Get("__response_")
	if result == nil {
		return nil
	}
	return result.(context.Response)
}

//ContextHandler api请求处理程序
func ContextHandler(handler servers.IExecuter, name string, engine string, service string, setting string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//处理输入参数
		mSetting, err := utility.GetMapWithQuery(setting)
		if err != nil {
			resp := context.GetStandardResponse()
			resp.SetError(500, err)
			setResponse(c, resp)
			return
		}
		ctx := context.GetContext(makeQueyStringData(c), makeFormData(c), makeParamsData(c), makeSettingData(c, mSetting), makeExtData(c))
		defer ctx.Close()
		defer setServiceName(c, ctx.Request.Translate(service, true))

		//调用执行引擎进行逻辑处理
		response, err := handler.Execute(name, engine, ctx.Request.Translate(service, true), ctx)
		if reflect.ValueOf(response).IsNil() {
			response = context.GetStandardResponse()
		}
		//处理错误err,5xx
		if err != nil {
			err = fmt.Errorf("error:%v", err)
			if !servers.IsDebug {
				err = errors.New("error:Internal Server Error(工作引擎发生异常)")
			}
			response.SetError(response.GetStatus(err), err)
			setResponse(c, response)
			return
		}

		//处理跳转3xx
		if url, ok := response.IsRedirect(); ok {
			c.Redirect(response.GetStatus(), url)
			return
		}

		//处理4xx,2xx
		setResponse(c, response)

	}
}

func makeFormData(ctx *gin.Context) InputData {
	return ctx.GetPostForm
}
func makeQueyStringData(ctx *gin.Context) InputData {
	return ctx.GetQuery
}
func makeParamsData(ctx *gin.Context) InputData {
	return ctx.Params.Get
}
func makeSettingData(ctx *gin.Context, m map[string]string) ParamData {
	return m
}

func makeExtData(c *gin.Context) map[string]interface{} {
	input := make(map[string]interface{})
	input["hydra_sid"] = getUUID(c)
	input["__jwt_"] = getJWTRaw(c)
	input["__func_http_request_"] = c.Request
	input["__func_http_response_"] = c.Request.Response
	input["__func_body_get_"] = func(ch string) (string, error) {
		if buff, ok := c.Get("__body_"); ok {
			return encoding.Convert(buff.([]byte), ch)
		}
		return "", errors.New("body读取错误")

	}
	return input
}

//InputData 输入参数
type InputData func(key string) (string, bool)

//Get 获取指定键对应的数据
func (i InputData) Get(key string) (string, error) {
	if r, ok := i(key); ok {
		return r, nil
	}
	return "", fmt.Errorf("数据不存在:%s", key)
}

//ParamData map参数数据
type ParamData map[string]string

//Get 获取指定键对应的数据
func (i ParamData) Get(key string) (string, error) {
	if r, ok := i[key]; ok {
		return r, nil
	}
	return "", fmt.Errorf("数据不存在:%s", key)
}
