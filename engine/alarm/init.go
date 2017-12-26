package alarm

import (
	"fmt"
	"strconv"

	"github.com/qxnw/hydra/context"

	"math"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

func (r *collectProxy) init() {
	r.reportSQL = `select * from alarm_records where "time"> now() - @time order by time desc`

	r.queryMap["http"] = `select value from alarm_records where "type"='http' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["tcp"] = `select value from alarm_records where "type"='tcp' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["registry"] = `select value from alarm_records where "type"='registry' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["db"] = `select value from alarm_records where "type"='db' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["cpu"] = `select value from alarm_records where "type"='cpu' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["mem"] = `select value from alarm_records where "type"='mem' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["disk"] = `select value from alarm_records where "type"='disk' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["ncc"] = `select value from alarm_records where "type"='ncc' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["nginx-error"] = `select value from alarm_records where "type"='nginx-error' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.queryMap["nginx-access"] = `select value from alarm_records where "type"='nginx-access' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`

	r.reportMap["http"] = "alarm_records type=http,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["tcp"] = "alarm_records type=tcp,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["registry"] = "alarm_records type=registry,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["db"] = "alarm_records type=db,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["cpu"] = "alarm_records type=cpu,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["mem"] = "alarm_records type=mem,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["disk"] = "alarm_records type=disk,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["ncc"] = "alarm_records type=ncc,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["nginx-error"] = "alarm_records type=nginx-error,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	r.reportMap["nginx-access"] = "alarm_records type=nginx-access,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	//服务器响应码
	r.srvQueryMap["api_server_reponse"] = `select m5 *300 as t from "api.server.response.meter" where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "url" fill(0) limit 1`
	r.queryMap["api_server_reponse"] = `select value from alarm_records where "type"='api_server_reponse' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["api_server_reponse"] = "alarm_records type=api_server_reponse,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	r.srvQueryMap["rpc_server_reponse"] = `select m5 *300 as t from "rpc.server.response.meter" where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "service" fill(0) limit 1`
	r.queryMap["rpc_server_reponse"] = `select value from alarm_records where "type"='rpc_server_reponse' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["rpc_server_reponse"] = "alarm_records type=rpc_server_reponse,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	r.srvQueryMap["web_server_reponse"] = `select m5 *300 as t from "web.server.response.meter" where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "url" fill(0) limit 1`
	r.queryMap["web_server_reponse"] = `select value from alarm_records where "type"='web_server_reponse' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["web_server_reponse"] = "alarm_records type=web_server_reponse,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	r.srvQueryMap["mq_consumer_reponse"] = `select m5 *300 as t from "mq.consumer.response.meter" where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "queue" fill(0) limit 1`
	r.queryMap["mq_consumer_reponse"] = `select value from alarm_records where "type"='mq_consumer_reponse' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["mq_consumer_reponse"] = "alarm_records type=mq_consumer_reponse,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	r.srvQueryMap["cron_server_reponse"] = `select m5 *300 as t from "cron.server.response.meter" where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "task" fill(0) limit 1`
	r.queryMap["cron_server_reponse"] = `select value from alarm_records where "type"='cron_server_reponse' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["cron_server_reponse"] = "alarm_records type=cron_server_reponse,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	//服务器并发数
	r.srvQueryMap["api_server_qps"] = `select m5 as t from "api.server.request.qps" where "domain" = '@domain' and "time" > now() - 5m group by "url" fill(0) limit 1`
	r.queryMap["api_server_qps"] = `select value from alarm_records where "type"='api_server_qps' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["api_server_qps"] = "alarm_records type=api_server_qps,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	r.srvQueryMap["web_server_qps"] = `select m5 as t from "web.server.request.qps" where "domain" = '@domain' and "time" > now() - 5m group by "url" fill(0) limit 1`
	r.queryMap["web_server_qps"] = `select value from alarm_records where "type"='web_server_qps' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["web_server_qps"] = "alarm_records type=web_server_qps,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	r.srvQueryMap["rpc_server_qps"] = `select m5 as t from "api.server.request.qps" where "domain" = '@domain' and "time" > now() - 5m group by "service" fill(0) limit 1`
	r.queryMap["rpc_server_qps"] = `select value from alarm_records where "type"='rpc_server_qps' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["rpc_server_qps"] = "alarm_records type=rpc_server_qps,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	r.srvQueryMap["mq_consumer_qps"] = `select m5 as t from "mq.consumer.request.qps" where "domain" = '@domain' and "time" > now() - 5m group by "queue" fill(0) limit 1`
	r.queryMap["mq_consumer_qps"] = `select value from alarm_records where "type"='mq_consumer_qps' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["mq_consumer_qps"] = "alarm_records type=mq_consumer_qps,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	r.srvQueryMap["job_server_qps"] = `select m5 as t from "job.server.request.qps" where "domain" = '@domain' and "time" > now() - 5m group by "task" fill(0) limit 1`
	r.queryMap["job_server_qps"] = `select value from alarm_records where "type"='job_server_qps' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	r.reportMap["job_server_qps"] = "alarm_records type=job_server_qps,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

}

func (s *collectProxy) query(ctx *context.Context, sql string, tf *transform.Transform) (domain []string, count []int, err error) {
	db, err := ctx.Influxdb.GetClient("metricdb")
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

var influxdbCache cmap.ConcurrentMap

func init() {
	influxdbCache = cmap.New(2)
}
