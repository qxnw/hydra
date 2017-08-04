package collect

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"

	"math"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/influxdb"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

func (r *collectProxy) init() {
	r.reportSQL = `select * from hydra_collector where "time">now() - @time order by time`

	r.queryMap["http"] = `select value from hydra_collector where "type"='http' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["tcp"] = `select value from hydra_collector where "type"='tcp' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["registry"] = `select value from hydra_collector where "type"='registry' and UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["db"] = `select value from hydra_collector where "type"='db' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["cpu"] = `select value from hydra_collector where "type"='cpu' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["mem"] = `select value from hydra_collector where "type"='mem' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["disk"] = `select value from hydra_collector where "type"='disk' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`

	r.reportMap["http"] = "hydra_collector,type=http,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["tcp"] = "hydra_collector,type=tcp,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["registry"] = "hydra_collector,type=registry,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg  value=@value"
	r.reportMap["db"] = "hydra_collector,type=db,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg  value=@value"
	r.reportMap["cpu"] = "hydra_collector,type=cpu,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg  value=@value"
	r.reportMap["mem"] = "hydra_collector,type=mem,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg  value=@value"
	r.reportMap["disk"] = "hydra_collector,type=disk,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg  value=@value"

	//服务器响应码
	r.srvQueryMap["api_server_reponse"] = `select m5 *300 as t from "api.server.response.meter"  where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "url" fill(0) limit 1`
	r.queryMap["api_server_reponse"] = `select value from hydra_collector where "type"='api_server_reponse' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["api_server_reponse"] = "hydra_collector,type=api_server_reponse,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg  value=@value"

	r.srvQueryMap["rpc_server_reponse"] = `select m5 *300 as t from "rpc.server.response.meter"  where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "service" fill(0) limit 1`
	r.queryMap["rpc_server_reponse"] = `select value from hydra_collector where "type"='rpc_server_reponse' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["rpc_server_reponse"] = "hydra_collector,type=rpc_server_reponse,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg  value=@value"

	r.srvQueryMap["web_server_reponse"] = `select m5 *300 as t from "web.server.response.meter"  where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "url" fill(0) limit 1`
	r.queryMap["web_server_reponse"] = `select value from hydra_collector where "type"='web_server_reponse' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["web_server_reponse"] = "hydra_collector,type=web_server_reponse,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg  value=@value"

	r.srvQueryMap["mq_consumer_reponse"] = `select m5 *300 as t from "mq.consumer.response.meter"  where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "queue" fill(0)  limit 1`
	r.queryMap["mq_consumer_reponse"] = `select value from hydra_collector where "type"='mq_consumer_reponse' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["mq_consumer_reponse"] = "hydra_collector,type=mq_consumer_reponse,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg  value=@value"

	r.srvQueryMap["cron_server_reponse"] = `select m5 *300 as t from "cron.server.response.meter"  where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "task" fill(0) limit 1`
	r.queryMap["cron_server_reponse"] = `select value from hydra_collector where "type"='cron_server_reponse' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["cron_server_reponse"] = "hydra_collector,type=cron_server_reponse,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg  value=@value"

	//服务器充值时长
	r.srvQueryMap["api_server_qps"] = `select sum(count) from "api.server.request.qps"  where "status" = '@code' and "time" > now() - @span  group by "domain"`
	r.queryMap["api_server_qps"] = `select value from hydra_collector where "type"='api_server_qps' and "domain"='@domain' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["api_server_qps"] = "hydra_collector,type=api_server_qps,domain=@domain,group=@group,level=@level,t=@time,msg=@msg  value=@value"

	//服务器并发数

}

func (s *collectProxy) query(ctx *context.Context, sql string, tf *transform.Transform) (domain []string, count []int, err error) {
	db, err := s.getInfluxClient(ctx, "metricdb")
	if err != nil {
		return
	}

	data, err := db.QueryResponse(sql)
	if err != nil {
		return
	}
	if err = data.Error(); err != nil {
		return
	}
	domain = make([]string, 0, 2)
	count = make([]int, 0, 2)
	for _, row := range data.Results {
		for _, ser := range row.Series {
			if len(ser.Tags) > 1 {
				err = fmt.Errorf("返回的数据集包含我个tag:%v", ser.Tags)
				return nil, nil, err
			}
			for _, v := range ser.Tags {
				domain = append(domain, v)
			}
			value, err := strconv.ParseFloat(types.GetString(ser.Values[0][1]), 64)
			if err != nil {
				err = fmt.Errorf("查询返回的数据不是数字:%v", data)
				return nil, nil, err
			}
			count = append(count, int(math.Floor(value)))
		}
	}
	return
}

var dbCache cmap.ConcurrentMap
var influxdbCache cmap.ConcurrentMap

func init() {
	dbCache = cmap.New(2)
	influxdbCache = cmap.New(2)
}
func (s *collectProxy) getInfluxClient(ctx *context.Context, name string) (*influxdb.InfluxClient, error) {
	content, err := ctx.GetVarParamByArgsName("influxdb", name)
	if err != nil {
		return nil, err
	}
	_, client, err := influxdbCache.SetIfAbsentCb(content, func(i ...interface{}) (interface{}, error) {
		cnf, err := conf.NewJSONConfWithJson(content, 0, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("args配置错误无法解析:%s(err:%v)", content, err)
		}
		host := cnf.String("host")
		dataBase := cnf.String("dataBase")
		if host == "" || dataBase == "" {
			return nil, fmt.Errorf("配置错误:host 和 dataBase不能为空（host:%s，dataBase:%s）", host, dataBase)
		}
		if !strings.Contains(host, "://") {
			host = "http://" + host
		}
		client, err := influxdb.NewInfluxClient(host, dataBase, cnf.String("userName"), cnf.String("password"))
		if err != nil {
			return nil, fmt.Errorf("初始化失败(err:%v)", err)
		}
		return client, err
	})
	if err != nil {
		return nil, err
	}
	return client.(*influxdb.InfluxClient), err

}

func (s *collectProxy) getDB(ctx *context.Context) (*db.DB, error) {
	db, err := ctx.GetArgByName("db")
	if err != nil {
		return nil, err
	}
	content, err := ctx.GetVarParam("db", db)
	if err != nil {
		return nil, fmt.Errorf("无法获取args参数db的值:%s(err:%v)", db, err)
	}
	return getDBFromCache(content)
}
func getDBFromCache(conf string) (*db.DB, error) {
	_, v, err := dbCache.SetIfAbsentCb(conf, func(input ...interface{}) (interface{}, error) {
		config := input[0].(string)
		configMap, err := jsons.Unmarshal([]byte(conf))
		if err != nil {
			return nil, err
		}
		provider, ok := configMap["provider"]
		if !ok {
			return nil, fmt.Errorf("db配置文件错误，未包含provider节点:%s", config)
		}
		connString, ok := configMap["connString"]
		if !ok {
			return nil, fmt.Errorf("db配置文件错误，未包含connString节点:%s", config)
		}
		return db.NewDB(provider.(string), connString.(string), types.ToInt(configMap["max"], 2))
	}, conf)
	if err != nil {
		return nil, err
	}
	return v.(*db.DB), nil
}
