package context

import (
	"fmt"

	"strings"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

//ContextDB 数据库操作
type ContextDB struct {
	ctx *Context
}

//Reset 重置context
func (cd *ContextDB) Reset(ctx *Context) {
	cd.ctx = ctx
}

//GetDB 获取数据库操作实例
func (cd *ContextDB) GetDB(names ...string) (d *db.DB, err error) {
	sName := "db"
	if len(names) > 0 {
		sName = names[0]
	}
	name, ok := cd.ctx.Input.Args[sName]
	if !ok {
		return nil, fmt.Errorf("未配置db参数(%v)", cd.ctx.Input.Args)
	}
	_, dbc, err := dbCache.SetIfAbsentCb(name, func(input ...interface{}) (d interface{}, err error) {
		name := input[0].(string)
		conf, err := cd.ctx.Input.GetVarParam("db", name)
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
		p, c, m := provider.(string), connString.(string), types.ToInt(configMap["max"], 2)
		d, err = db.NewDB(p, c, m)
		if err != nil {
			err = fmt.Errorf("创建DB失败:%s,%s,%d,err:%v", p, c, m, err)
			return
		}
		return
	}, name)
	if err != nil {
		return
	}
	d = dbc.(*db.DB)
	return
}

//Scalar 获取首行首列值
func (cd *ContextDB) Scalar(tpl []string, input map[string]interface{}) (data interface{}, err error) {
	db, err := cd.GetDB()
	if err != nil {
		return
	}
	sql, _, err := cd.getSQL(tpl)
	if err != nil {
		return
	}
	data, _, _, err = db.Scalar(sql, input)
	if err != nil {
		err = fmt.Errorf("执行scalar失败：%s,%v,err:%v", sql, input, err)
		return
	}
	return
}

//Execute  执行数据库操作
func (cd *ContextDB) Execute(tpl []string, input map[string]interface{}) (row int64, err error) {
	db, err := cd.GetDB()
	if err != nil {
		return
	}
	sql, key, err := cd.getSQL(tpl)
	if err != nil {
		return
	}
	row, _, _, err = db.Execute(sql, input)
	if err != nil {
		err = fmt.Errorf("执行SQL语句失败:%s,:%v,err:%v", sql, input, err)
		return
	}
	if len(key) > 0 {
		c, err := cd.ctx.GetCache()
		if err != nil {
			cd.ctx.Error("清除缓存,获取缓存操作实例失败:", err)
			return row, nil
		}
		tf := transform.NewMaps(input)
		for _, v := range key {
			if v == "" {
				continue
			}
			err = c.Delete(tf.Translate(v))
			if err != nil {
				cd.ctx.Errorf("清除缓存失败：%s,%v", v, err)
			}
		}

	}
	return
}

//getSQL 获取SQL语句
func (cd *ContextDB) getSQL(tpl []string) (sql string, key []string, err error) {
	if len(tpl) < 1 {
		err = fmt.Errorf("输入的SQL模板错误，必须包含1个元素(SQL语句):%v", tpl)
		return
	}
	sql = tpl[0]
	if len(tpl) > 1 {
		key = strings.Split(tpl[1], ";")
	}
	return
}

//GetFirstRow 获取首行数据，数据不存在时返回ErrDataNotExist错误
func (cd *ContextDB) GetFirstRow(tpl []string, input map[string]interface{}) (data db.QueryRow, err error) {
	result, err := cd.GetDataRows(tpl, input)
	if err != nil {
		return
	}
	if len(result) > 0 {
		return result[0], nil
	}
	return nil, ERR_DataNotExist
}

//GetDataRows 获取多行数据
func (cd *ContextDB) GetDataRows(tpl []string, input map[string]interface{}) (data []db.QueryRow, err error) {
	db, err := cd.GetDB()
	if err != nil {
		err = fmt.Errorf("获取数据库操作实例失败:err:%v", err)
		return
	}
	sql, _, err := cd.getSQL(tpl)
	if err != nil {
		return
	}
	data, _, _, err = db.Query(sql, input)
	if err != nil {
		err = fmt.Errorf("执行查询发生错误:%s,%v,err:%v", sql, input, err)
	}
	return
}

var dbCache cmap.ConcurrentMap

func init() {
	dbCache = cmap.New(2)
}
