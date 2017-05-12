package influxdb

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/transform"
)

type influxProxy struct {
	domain     string
	serverName string
	serverType string
	services   []string
	dbs        cmap.ConcurrentMap
}

func newInfluxProxy() *influxProxy {
	return &influxProxy{
		services: make([]string, 0, 16),
		dbs:      cmap.New(),
	}
}

func (s *influxProxy) Start(domain string, serverName string, serverType string, invoker *rpc.RPCInvoker) (services []string, err error) {
	s.domain = domain
	s.serverName = serverName
	s.serverType = serverType
	return []string{}, nil

}
func (s *influxProxy) Close() error {
	s.dbs.RemoveIterCb(func(key string, value interface{}) bool {
		client := value.(*influxClientConf)
		client.client.Close()
		return true
	})
	return nil
}
func (s *influxProxy) getHandleParams(ctx *context.Context) (*transform.Transform, error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil || ctx.Input.Params == nil {
		return nil, fmt.Errorf("input,params,args不能为空:%v", ctx.Input)
	}

	input, ok := ctx.Input.Input.(transform.ITransformGetter)
	if !ok {
		return nil, fmt.Errorf("input输入参数类型错误不是ITransformGetter类型:%v", ctx.Input.Args)
	}
	params, ok := ctx.Input.Params.(transform.ITransformGetter)
	if !ok {
		return nil, fmt.Errorf("params输入参数类型错误不是ITransformGetter类型:%v", ctx.Input.Args)
	}
	params.Set("domain", s.domain)
	tfParams := transform.NewGetter(params)
	tfParams.Append(input)
	return tfParams, nil

}
func (s *influxProxy) getDbClient(ctx *context.Context) (*influxClientConf, error) {
	argsMap, ok := ctx.Input.Args.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("Args输入参数类型错误不是map[string]string类型:%v", ctx.Input.Args)
	}
	db, ok := argsMap["db"]
	if db == "" || !ok {
		return nil, fmt.Errorf("engine:influxdb.args配置错误，缺少db参数:%v", ctx.Input.Args)
	}
	content, err := s.getVarParam(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("engine:无法获取args参数db的值:%s(err:%v)", db, err)
	}
	client, err := s.getInfluxClient(content)
	if err != nil {
		return nil, fmt.Errorf("engine:无法创建influxdb:%s(err:%v)", content, err)
	}
	return client, nil

}
func (s *influxProxy) insert(client *influxClientConf, tfParams *transform.Transform) (err error) {
	//翻译并转换参数
	measurement := tfParams.Translate(client.measurement)
	tags := make(map[string]string)
	for k, v := range client.tags {
		key := tfParams.Translate(k)
		tags[key] = tfParams.Translate(v)
	}
	fileds := make(map[string]interface{})
	for k, v := range client.fields {
		key := tfParams.Translate(k)
		fileds[key] = tfParams.Translate(v)
	}
	err = client.client.Send(measurement, tags, fileds)
	if err != nil {
		return fmt.Errorf("engine:消息发送到influxdb失败:(err:%v)", err)
	}
	return nil
}
func (s *influxProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {
	//获取基础参数
	tfParams, err := s.getHandleParams(ctx)
	if err != nil {
		return &context.Response{Status: 500}, err
	}

	//获取influxdb
	client, err := s.getDbClient(ctx)
	if err != nil {
		return &context.Response{Status: 500}, err
	}
	method, err := tfParams.Get("method")
	if err != nil {
		return &context.Response{Status: 500}, err
	}
	switch method {
	case "GET":
	case "REQUEST":
	case "POST":
	case "INSERT":
		err = s.insert(client, tfParams)
		if err != nil {
			return &context.Response{Status: 500}, err
		}
	}

	return &context.Response{Status: 200, Content: "SUCCESS"}, nil
}
func (s *influxProxy) getInfluxClient(content string) (*influxClientConf, error) {
	_, client, err := s.dbs.SetIfAbsentCb(content, func(i ...interface{}) (interface{}, error) {
		cnf, err := conf.NewJSONConfWithJson(content, 0, nil)
		if err != nil {
			return nil, fmt.Errorf("engine:influxdb.args配置错误无法解析:%s(err:%v)", content, err)
		}
		host := cnf.String("host")
		dataBase := cnf.String("dataBase")
		if host == "" || dataBase == "" {
			return nil, fmt.Errorf("engine:influxdb配置错误:host 和 dataBase不能为空（host:%s，dataBase:%s）", host, dataBase)
		}
		if !strings.Contains(host, "://") {
			host = "http://" + host
		}
		client, err := newInfluxClient(host, dataBase, cnf.String("userName"), cnf.String("password"))
		if err != nil {
			return nil, fmt.Errorf("engine:influxdb初始化失败(err:%v)", err)
		}
		engineSection, err := cnf.GetSection("engine")
		if err != nil {
			return nil, fmt.Errorf("engine:配置错误:%s(err:%v)", content, err)
		}
		clientConf := &influxClientConf{}
		clientConf.client = client
		clientConf.measurement = engineSection.String("measurement")
		if clientConf.measurement == "" {
			return nil, fmt.Errorf("engine:配置错误:%s(err:engine.measurement未配置)", content)
		}
		clientConf.fields, err = engineSection.GetSMap("fields")
		if err != nil {
			return nil, err
		}
		clientConf.tags, err = engineSection.GetSMap("tags")
		if err != nil {
			return nil, err
		}
		return clientConf, err
	})
	if err != nil {
		return nil, err
	}
	return client.(*influxClientConf), err

}
func (s *influxProxy) getVarParam(ctx *context.Context, name string) (string, error) {
	func_var := ctx.Ext["__func_var_get_"]
	if func_var == nil {
		return "", errors.New("engine:未找到__func_var_get_")
	}
	if f, ok := func_var.(func(c string, n string) (string, error)); ok {
		return f("db", name)
	}
	return "", errors.New("engine:未找到__func_var_get_传入类型错误")
}

func (s *influxProxy) Has(shortName, fullName string) (err error) {
	return nil
}

type influxProxyResolver struct {
}

func (s *influxProxyResolver) Resolve() engine.IWorker {
	return newInfluxProxy()
}

func init() {
	engine.Register("influx", &influxProxyResolver{})
}
