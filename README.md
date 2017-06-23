# hydra 
通用服务框架，实现一个框架集成多个服务如：http server,rpc server,mq consumer,job server等,并提供一致的业务监控，日志存储，服务注册与发现，负载均衡，限流，限量，自动更新等统一治理可通过go或lua编写业务代码，并集成:缓存服务，数据库服务，邮件服务，短信服务，RCP代理,MOCK等基础服务

  hydra特点
* 部署简单: 本地零配置，通过服务治理中心简单配置即可启动; 每台服务器可部署多个hydra,每个hydra可启动多个服务
* 业务扩展: 业务代码可通过lua,go快速开发，只需关注业务本身，简单配置便可运行于任何服务器
* 智能监控: 自动监控QPS, 服务执行时长，执行结果，内存，CPU，服务运行状态并自动上报到influxdb
* 统一日志: 可统一输出到kafka或本地文件，对请求加入sid,可根据sid查询到该业务的多个服务器执行日志
* 集成引擎: 集成缓存服务，数据库服务，邮件服务，短信服务，RCP代理,Http代理，MOCK等简单配置即可使用
* 服务治理: 通过zookeeper注册与发现服务，多种负载均衡，流量控制，灰度发布等
* 简单高效: http api以tango，rpc以grpc,cron以timewheel为基础进行开发，逻辑简单可应付大流量场景




## 下载 安装

    go get github.com/qxnw/hydra


    