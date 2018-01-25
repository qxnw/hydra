package alarm

var queryMap = make(map[string]string)
var reportMap = make(map[string]string)
var srvQueryMap = make(map[string]string)
var reportSQL string

func init() {

	reportSQL = `select * from alarm_records where "time"> now() - @time order by time desc`

	queryMap["http"] = `select value from alarm_records where "type"='http' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	queryMap["tcp"] = `select value from alarm_records where "type"='tcp' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	queryMap["registry"] = `select value from alarm_records where "type"='registry' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	queryMap["db"] = `select value from alarm_records where "type"='db' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	queryMap["cpu"] = `select value from alarm_records where "type"='cpu' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	queryMap["mem"] = `select value from alarm_records where "type"='mem' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	queryMap["disk"] = `select value from alarm_records where "type"='disk' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	queryMap["ncc"] = `select value from alarm_records where "type"='ncc' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	queryMap["nginx-error"] = `select value from alarm_records where "type"='nginx-error' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	queryMap["nginx-access"] = `select value from alarm_records where "type"='nginx-access' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	queryMap["queue-count"] = `select value from alarm_records where "type"='queue-count' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`

	reportMap["http"] = "alarm_records type=http,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	reportMap["tcp"] = "alarm_records type=tcp,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	reportMap["registry"] = "alarm_records type=registry,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	reportMap["db"] = "alarm_records type=db,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	reportMap["cpu"] = "alarm_records type=cpu,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	reportMap["mem"] = "alarm_records type=mem,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	reportMap["disk"] = "alarm_records type=disk,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	reportMap["ncc"] = "alarm_records type=ncc,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	reportMap["nginx-error"] = "alarm_records type=nginx-error,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	reportMap["nginx-access"] = "alarm_records type=nginx-access,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"
	reportMap["queue-count"] = "alarm_records type=queue-count,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	//服务器响应码
	srvQueryMap["api_server_reponse"] = `select m5 *300 as t from "api.server.response.meter" where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "url" fill(0) limit 1`
	queryMap["api_server_reponse"] = `select value from alarm_records where "type"='api_server_reponse' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	reportMap["api_server_reponse"] = "alarm_records type=api_server_reponse,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	srvQueryMap["rpc_server_reponse"] = `select m5 *300 as t from "rpc.server.response.meter" where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "service" fill(0) limit 1`
	queryMap["rpc_server_reponse"] = `select value from alarm_records where "type"='rpc_server_reponse' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	reportMap["rpc_server_reponse"] = "alarm_records type=rpc_server_reponse,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	srvQueryMap["web_server_reponse"] = `select m5 *300 as t from "web.server.response.meter" where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "url" fill(0) limit 1`
	queryMap["web_server_reponse"] = `select value from alarm_records where "type"='web_server_reponse' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	reportMap["web_server_reponse"] = "alarm_records type=web_server_reponse,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	srvQueryMap["mq_consumer_reponse"] = `select m5 *300 as t from "mq.consumer.response.meter" where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "queue" fill(0) limit 1`
	queryMap["mq_consumer_reponse"] = `select value from alarm_records where "type"='mq_consumer_reponse' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	reportMap["mq_consumer_reponse"] = "alarm_records type=mq_consumer_reponse,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	srvQueryMap["cron_server_reponse"] = `select m5 *300 as t from "cron.server.response.meter" where "domain" = '@domain' and "status" = '@code' and "time" > now() - 5m group by "task" fill(0) limit 1`
	queryMap["cron_server_reponse"] = `select value from alarm_records where "type"='cron_server_reponse' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	reportMap["cron_server_reponse"] = "alarm_records type=cron_server_reponse,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	//服务器并发数
	srvQueryMap["api_server_qps"] = `select m5 as t from "api.server.request.qps" where "domain" = '@domain' and "time" > now() - 5m group by "url" fill(0) limit 1`
	queryMap["api_server_qps"] = `select value from alarm_records where "type"='api_server_qps' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	reportMap["api_server_qps"] = "alarm_records type=api_server_qps,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	srvQueryMap["web_server_qps"] = `select m5 as t from "web.server.request.qps" where "domain" = '@domain' and "time" > now() - 5m group by "url" fill(0) limit 1`
	queryMap["web_server_qps"] = `select value from alarm_records where "type"='web_server_qps' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	reportMap["web_server_qps"] = "alarm_records type=web_server_qps,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	srvQueryMap["rpc_server_qps"] = `select m5 as t from "api.server.request.qps" where "domain" = '@domain' and "time" > now() - 5m group by "service" fill(0) limit 1`
	queryMap["rpc_server_qps"] = `select value from alarm_records where "type"='rpc_server_qps' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	reportMap["rpc_server_qps"] = "alarm_records type=rpc_server_qps,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	srvQueryMap["mq_consumer_qps"] = `select m5 as t from "mq.consumer.request.qps" where "domain" = '@domain' and "time" > now() - 5m group by "queue" fill(0) limit 1`
	queryMap["mq_consumer_qps"] = `select value from alarm_records where "type"='mq_consumer_qps' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	reportMap["mq_consumer_qps"] = "alarm_records type=mq_consumer_qps,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

	srvQueryMap["job_server_qps"] = `select m5 as t from "job.server.request.qps" where "domain" = '@domain' and "time" > now() - 5m group by "task" fill(0) limit 1`
	queryMap["job_server_qps"] = `select value from alarm_records where "type"='job_server_qps' and platform='@platform' and "UNQ"='@unq' and "time">'now()-6h' order by time desc limit 1`
	reportMap["job_server_qps"] = "alarm_records type=job_server_qps,platform=@platform,UNQ=@unq,title=@title,group=@group,level=@level,t=@time,msg=@msg value=@value"

}
