package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/qxnw/hydra_plugin/plugins"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/memcache"
	"github.com/qxnw/lib4go/transform"
)

var contextPool *sync.Pool

func init() {
	contextPool = &sync.Pool{
		New: func() interface{} {
			return &wxContext{}
		},
	}
}

type wxContext struct {
	service      string
	ctx          plugins.Context
	Input        transform.ITransformGetter
	Params       transform.ITransformGetter
	Body         string
	db           *db.DB
	Args         map[string]string
	func_var_get func(c string, n string) (string, error)
	RPC          plugins.RPCInvoker
	*logger.Logger
}

func (w *wxContext) CheckMustFields(names ...string) error {
	for _, v := range names {
		if _, err := w.Input.Get(v); err != nil {
			err := fmt.Errorf("输入参数:%s不能为空", v)
			return err
		}
	}
	return nil
}

func GetWXContext(ctx plugins.Context, invoker plugins.RPCInvoker) (wx *wxContext, err error) {
	wx = contextPool.Get().(*wxContext)
	wx.ctx = ctx
	defer func() {
		if err != nil {
			wx.Close()
		}
	}()
	wx.Input, err = wx.getGetParams(ctx.GetInput())
	if err != nil {
		return
	}
	wx.Params, err = wx.getGetParams(ctx.GetParams())
	if err != nil {
		return
	}
	wx.Body, err = wx.getGetBody(ctx.GetBody())
	if err != nil {
		return
	}
	wx.Args, err = wx.GetArgs(ctx.GetArgs())
	if err != nil {
		return
	}
	wx.func_var_get, err = wx.getVarParam(ctx.GetExt())
	if err != nil {
		return
	}
	wx.Logger, err = wx.getLogger()
	if err != nil {
		return
	}
	wx.RPC = invoker
	return
}

func (w *wxContext) GetCache() (c *memcache.MemcacheClient, err error) {
	name, ok := w.Args["cache"]
	if !ok {
		return nil, fmt.Errorf("服务%s未配置cache参数(%v)", w.service, w.Args)
	}
	conf, err := w.func_var_get("cache", name)
	if err != nil {
		return nil, err
	}
	configMap, err := jsons.Unmarshal([]byte(conf))
	if err != nil {
		return nil, err
	}
	server, ok := configMap["server"]
	if !ok {
		err = fmt.Errorf("cache[%s]配置文件错误，未包含server节点:%s", name, conf)
		return nil, err
	}
	return memcache.New(strings.Split(server.(string), ";"))

}

func (w *wxContext) GetJsonFromCache(sql string, input map[string]interface{}) (cvalue string, err error) {
	db, err := w.GetDB()
	if err != nil {
		return
	}
	query, params := db.GetTPL().GetSQLContext(sql, input)
	key := fmt.Sprintf("%s:%+v", query, params)
	client, err := w.GetCache()
	if err != nil {
		return
	}
	cvalue, err = client.Get(key)
	if err != nil {
		return
	}
	if cvalue != "" {
		return
	}
	data, _, _, err := db.Query(sql, input)
	if err != nil {
		return
	}
	buffer, err := jsons.Marshal(&data)
	if err != nil {
		return
	}
	client.Set(key, string(buffer), 0)
	return
}
func (w *wxContext) GetFirstMapFromCache(sql string, input map[string]interface{}) (data map[string]interface{}, err error) {
	result, err := w.GetMapFromCache(sql, input)
	if err != nil {
		return
	}
	if len(result) > 0 {
		return result[0], nil
	}
	return nil, fmt.Errorf("返回的数据条数为0:(%s)", sql)

}
func (w *wxContext) GetMapFromCache(sql string, input map[string]interface{}) (data []map[string]interface{}, err error) {
	db, err := w.GetDB()
	if err != nil {
		return
	}
	query, params := db.GetTPL().GetSQLContext(sql, input)
	key := fmt.Sprintf("%s:%+v", query, params)
	client, err := w.GetCache()
	if err != nil {
		return
	}
	dstr, err := client.Get(key)
	if err != nil {
		return
	}
	if dstr != "" {
		err = json.Unmarshal([]byte(dstr), &data)
		return
	}
	data, _, _, err = db.Query(sql, input)
	if err != nil {
		return
	}
	cvalue, err := jsons.Marshal(data)
	if err != nil {
		return
	}
	client.Set(key, string(cvalue), 0)
	return
}
func (w *wxContext) ScalarFromDb(sql string, input map[string]interface{}) (data interface{}, err error) {
	db, err := w.GetDB()
	if err != nil {
		return
	}
	data, _, _, err = db.Scalar(sql, input)
	return
}
func (w *wxContext) ExecuteToDb(sql string, input map[string]interface{}) (row int64, err error) {
	db, err := w.GetDB()
	if err != nil {
		return
	}
	row, _, _, err = db.Execute(sql, input)
	return
}
func (w *wxContext) GetDataFromDb(sql string, input map[string]interface{}) (data []map[string]interface{}, err error) {
	db, err := w.GetDB()
	if err != nil {
		return
	}
	data, _, _, err = db.Query(sql, input)
	return
}

func (w *wxContext) GetDB() (d *db.DB, err error) {
	if w.db != nil {
		return w.db, nil
	}
	name, ok := w.Args["db"]
	if !ok {
		return nil, fmt.Errorf("服务%s未配置db参数(%v)", w.service, w.Args)
	}
	conf, err := w.func_var_get("db", name)
	if err != nil {
		return nil, err
	}
	configMap, err := jsons.Unmarshal([]byte(conf))
	if err != nil {
		return nil, err
	}
	provider, ok := configMap["provider"]
	if !ok {
		return nil, fmt.Errorf("db配置文件错误，未包含provider节点:var/db/%s", name)
	}
	connString, ok := configMap["connString"]
	if !ok {
		return nil, fmt.Errorf("db配置文件错误，未包含connString节点:var/db/%s", name)
	}
	d, err = db.NewDB(provider.(string), connString.(string), 2)
	if err != nil {
		err = fmt.Errorf("创建DB失败:err:%v", err)
		return
	}
	w.db = d
	return
}

func (w *wxContext) getLogger() (*logger.Logger, error) {
	if session, ok := w.ctx.GetExt()["hydra_sid"]; ok {
		return logger.GetSession("wx_base_core", session.(string)), nil
	}
	return nil, fmt.Errorf("输入的context里没有包含hydra_sid(%v)", w.ctx.GetExt())
}
func (w *wxContext) getVarParam(ext map[string]interface{}) (func(c string, n string) (string, error), error) {
	funcVar := ext["__func_var_get_"]
	if funcVar == nil {
		return nil, errors.New("未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f, nil
	}
	return nil, errors.New("未找到__func_var_get_传入类型错误")
}
func (w *wxContext) GetArgs(args interface{}) (params map[string]string, err error) {
	params, ok := args.(map[string]string)
	if !ok {
		err = fmt.Errorf("未设置Args参数")
		return
	}
	return
}
func (w *wxContext) getGetBody(body interface{}) (t string, err error) {
	if body == nil {
		return "", errors.New("body 数据为空")
	}
	t, ok := body.(string)
	if !ok {
		return "", errors.New("body 不是字符串数据")
	}
	return
}
func (w *wxContext) getGetParams(input interface{}) (t transform.ITransformGetter, err error) {
	if input == nil {
		err = fmt.Errorf("输入参数为空:%v", input)
		return nil, err
	}
	t, ok := input.(transform.ITransformGetter)
	if !ok {
		return t, fmt.Errorf("输入参数为空:input（%v）不是transform.ITransformGetter类型", input)
	}
	return t, nil
}
func (w *wxContext) Close() {
	contextPool.Put(w)
}
