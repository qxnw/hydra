package memcache

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engine"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/memcache"
	"github.com/qxnw/lib4go/transform"
)

type memcacheProxy struct {
	domain     string
	serverName string
	serverType string
	dbs        cmap.ConcurrentMap
}

func newMemcacheProxy() *memcacheProxy {
	return &memcacheProxy{
		dbs: cmap.New(),
	}
}

func (s *memcacheProxy) Start(domain string, serverName string, serverType string, invoker *rpc.RPCInvoker) (services []string, err error) {
	s.domain = domain
	s.serverName = serverName
	s.serverType = serverType
	return []string{}, nil

}
func (s *memcacheProxy) Close() error {
	return nil
}

func (s *memcacheProxy) getHandleParams(ctx *context.Context) (*transform.Transform, error) {
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
func (s *memcacheProxy) Handle(svName string, mode string, service string, ctx *context.Context) (r *context.Response, err error) {

	tfParams, err := s.getHandleParams(ctx)
	if err != nil {
		return &context.Response{Status: 500}, err
	}

	args, err := s.getArgs(ctx)
	if err != nil {
		return nil, err
	}
	db, ok := args["db"]
	if db == "" || !ok {
		return nil, fmt.Errorf("engine:influxdb.args配置错误，缺少db参数:%v", ctx.Input.Args)
	}
	content, err := s.getVarParam(ctx, db)
	if err != nil {
		return nil, err
	}
	client, err := s.getMemcacheClient(content)
	if err != nil {
		return nil, err
	}
	method, err := tfParams.Get("method")
	if err != nil {
		return &context.Response{Status: 500}, err
	}
	key, err := tfParams.Get("key")
	if err != nil {
		return &context.Response{Status: 500}, err
	}
	switch method {
	case "GET":
	case "QUERY":
		result := client.Get(key)
		return &context.Response{Status: 200, Content: result}, nil
	case "REQUEST":
	case "POST":
	case "INSERT":
		err := client.Set(key, "", 0)
		if err != nil {
			return &context.Response{Status: 500}, err
		}
		return &context.Response{Status: 200}, nil
	}
	return &context.Response{Status: 500}, err
}

func (s *memcacheProxy) getMemcacheClient(content string) (*memcache.MemcacheClient, error) {
	_, client, err := s.dbs.SetIfAbsentCb(content, func(i ...interface{}) (interface{}, error) {
		cnf, err := conf.NewJSONConfWithJson(content, 0, nil)
		if err != nil {
			return nil, fmt.Errorf("engine:influxdb.args配置错误无法解析:%s(err:%v)", content, err)
		}
		host := cnf.Strings("host")
		if len(host) == 0 {
			return nil, fmt.Errorf("engine:influxdb配置错误:host 和 dataBase不能为空（host:%v）", host)
		}
		mem, err := memcache.New(host)
		return mem, err
	})
	if err != nil {
		return nil, err
	}
	return client.(*memcache.MemcacheClient), err

}
func (s *memcacheProxy) Has(shortName, fullName string) (err error) {
	return nil
}
func (s *memcacheProxy) getArgs(ctx *context.Context) (map[string]string, error) {
	argsMap, ok := ctx.Input.Args.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("Args输入参数类型错误不是map[string]string类型:%v", ctx.Input.Args)
	}
	return argsMap, nil
}
func (s *memcacheProxy) getVarParam(ctx *context.Context, name string) (string, error) {
	func_var := ctx.Ext["__func_var_get_"]
	if func_var == nil {
		return "", errors.New("engine:未找到__func_var_get_")
	}
	if f, ok := func_var.(func(c string, n string) (string, error)); ok {
		return f("db", name)
	}
	return "", errors.New("engine:未找到__func_var_get_传入类型错误")
}

type memcacheProxyResolver struct {
}

func (s *memcacheProxyResolver) Resolve() engine.IWorker {
	return newMemcacheProxy()
}

func init() {
	engine.Register("memcache", &memcacheProxyResolver{})
}
