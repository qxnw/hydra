# hydra 
通用服务框架，实现一个框架集成多个服务器如：http server,rpc server,mq consumer,job server等,并提供一致的业务监控，日志存储，服务注册与发现，负载均衡，限流，限量，自动更新等统一治理可通过go或lua编写业务代码，并集成:缓存服务，数据库服务，邮件服务，短信服务，RCP代理,MOCK等基础服务引擎简单配置即可挂载到任意服务器

  hydra特点
* 部署简单: 本地零配置，通过服务治理中心简单配置即可启动; 每台服务器可部署多个hydra,每个hydra可启动多个服务,可通过配置启动或停止任意服务
* 业务扩展: 业务代码可通过lua,go快速开发，只需关注业务本身，简单配置便可运行于任何服务器
* 智能监控: 自动监控QPS, 服务执行时长，执行结果，内存，CPU，服务运行状态并自动上报到influxdb
* 统一日志: 可统一输出到kafka或本地文件，对请求加入sid,可根据sid查询到该业务的多个服务器执行日志
* 集成引擎: 集成缓存服务，数据库服务，邮件服务，短信服务，RCP代理,Http代理，MOCK等简单配置即可使用
* 服务治理: 通过zookeeper注册与发现服务，多种负载均衡，流量控制，灰度发布等
* 简单高效: http api以tango，rpc以grpc,cron以timewheel为基础进行开发，逻辑简单可应付大流量场景

 # 服务器介绍
   
  #### http api server
* 支持RESTful,适合以http接口作为服务的场景，将go,lua或集成服务，配置到http路由配置中即可供外部调用
* 服务器监控地址，路由变化后自动重启，不影响业务。非关健配置变更无需重启服务直接更新
* 只需简单配置可作为RPC服务转换为HTTP服务直接调用
* 固定返回结果的请求配置为mock模式用于测试打桩
 

 
  #### rpc server 
* 基于grpc实现
* 多种负载均衡轮询，权重，本地优先等
* 支持同步，异步，并行调用,支持快速失败，支持限流请求
 


  mq consumer
* 可监控多个queue消息，收到消息后可通过任意引擎执行
* 自动重连mq server无需重启服务器



   cron server
* 基于timewheel算法，可支持任意多的job
* 可配置固定参数发起http请求，RPC等任意服务请求
* 支持cron表达式，可配置任意执行时间
* 支持集群部署，每次只有一台机器执行







## 下载 安装

    go get github.com/qxnw/hydra


