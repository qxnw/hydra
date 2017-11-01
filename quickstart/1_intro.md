# 快速入门
hydra是一个能够快速开发http接口, web应用，RPC服务，流程服务，定时任务，消息消费(MQ Consumer)的微服务应用框架，你只需要关注你所提供的服务自身，使用注册中心简单的配置并通过hydra启动起来即可对外提供服务。


  hydra特点
* 部署简单: 打包hydra和业务动态库，复制到服务器通过命令启动起来即可; 
* 本地零配置: 只需指定注册中心地址和平台名称，启动后自动从注册中心拉取平台配置，配置变更后自动更新服务器，必要时自动优雅重启服务器; 
* 扩展简单: 业务代码通过go插件编写，实现1个接口生成动态库即可。开发者只需关注所提供的服务本身，开发的服务可作为http接口，Web应用，RPC服务，消息处理，定时任务等运行
* 智能监控: 自动统计QPS, 服务执行时长，执行结果，机器的CPU，内存等并自动上报到influxdb，通过grafana配置后即可看到动态图表
* 统一日志:请求自动生成UUID,跨服务器请求时也自动传入UUID,通过UUID可查询到同一请求的所有执行日志; 并集成RPC日志，系统自动将日志上传到RPC服务器，通过elasticsearch存储，使用themis即可查看日志内容
* 内置引擎: 资源(http,tcp,registry,cpu,memory,disk,db,net)状态检测(monitor)与报警(alarm),文件上传，mock,缓存，短信发送,微信消息推送，RPC服务代理等，通过简单配置即可实现如报警监控，动态图表，文件上传服务器，消息发送服务器，接口mock测试等
* 服务治理: 使用themis管理服务器配置如：安全认证，负载均衡，流量控制，灰度发布等
* 简单高效: 只需实现1个接口，简单配置即可运行。开发效率高。使用go原生服务器为基础进行扩展，可支持高并发
* 混合服务：同一个hydra可运行多个服务器，支持的服务器有:http接口服务器，web服务器，RPC服务器，mq consumer,自动任务5种服务器

[hydra安装](https://github.com/qxnw/hydra/blob/master/quickstart/2_install.md)

[gaea工具简介](https://github.com/qxnw/hydra/blob/master/quickstart/3.install_gaea.md)



## hydra架构图

![架构图](https://github.com/qxnw/hydra/blob/master/quickstart/hydra.png?raw=true)


## hydra启动过程


![架构图](https://github.com/qxnw/hydra/blob/master/quickstart/flow.png?raw=true)