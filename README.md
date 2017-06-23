# hydra 
通用业务服务框架，实现框架集成多个常用服务如：http,rpc,mq,job等,并提供一致的业务监控，日志存储，服务注册与发现，负载均衡，限流，限量，自动更新等统一治理
同时集中多种引擎如:缓存服务，数据库服务，邮件服务，短信服务，RCP代理,MOCK等

#hydra特点
* 简单化：本地零配置，注册中心简单配置即可启动服务。业务代码可通过lua,go快速开发，通过hydra工具一键发布。集成常用执行引擎(缓存,数据库，邮件等)，直接作为服务调用。
* 智能化: 支持智能路由，智能监控可监控QPS,执行时长，执行结果，内存，CPU，服务运行状态并可自动上报到influxdb。
* 高性能： http api以tango，rpc以grpc,cron以timewheel为基础进行开发，逻辑简单可应付大流量场景



