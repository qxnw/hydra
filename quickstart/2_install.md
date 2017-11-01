# hydra的安装

hydra源码的下载和安装可使用go get命令，由于hydra使用了较多的外部源码,此过程可能消耗较长时间

* 初次安装hydra
```
go get github.com/qxnw/hydra
```

* 更新hydra

```
go get -u github.com/qxnw/hydra
```
## hydra相关组件介绍

| 外部组件        | 必须安装           | 说明  |
| ------------- |:-------------:| -----|
|注册中心    | 是 |必须安装，用于管理服务器配置，目前只支持zookeeper|
|themis|否|建议安装，服务器配置提供图形化界面方便操作|
|gaea|否|建议安装，创建或管理hydra项目，可提高开发效率|
|oci|否|开发基于oracle数据库功能时安装|
|influxdb    | 否|   需要收集服务器监控数据时安装 |
|stomp mq |否| 开发mq consumer时安装 |
|elasticsearch|否|需要使用统一日志收集功能时安装|
|memcached|否|需要使用memcached功能时安装|
|redis|否|需要使用redis功能时安装|



