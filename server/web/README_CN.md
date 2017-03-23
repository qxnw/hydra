WebServer [![Build Status](https://drone.io/github.com/lunny/WebServer/status.png)](https://drone.io/github.com/lunny/WebServer/latest) [![](http://gocover.io/_badge/github.com/lunny/WebServer)](http://gocover.io/github.com/lunny/WebServer) [English](README.md)
=======================

![WebServer Logo](logo.png)

WebServer 是一个微内核的Go语言Web框架，采用模块化和注入式的设计理念。开发者可根据自身业务逻辑来选择性的装卸框架的功能，甚至利用丰富的中间件来搭建一个全栈式Web开发框架。

## 最近更新
- [2016-5-12] 开放Route级别中间件支持
- [2016-3-16] Group完善中间件支持，Route支持中间件
- [2016-2-1] 新增 session-ssdb，支持将ssdb作为session的后端存储
- [2015-10-23] 更新[renders](https://github.com/WebServer-contrib/renders)插件，解决模板修改后需要刷新两次才能生效的问题

## 特性
- 强大而灵活的路由设计
- 兼容已有的`http.Handler`
- 基于中间件的模块化设计，灵活定制框架功能
- 高性能的依赖注入方式

## 安装WebServer：
    go get github.com/lunny/WebServer

## 快速入门

一个经典的WebServer例子如下：

```go
package main

import (
    "errors"
    "github.com/lunny/WebServer"
)

type Action struct {
    WebServer.Json
}

func (Action) Get() interface{} {
    if true {
        return map[string]string{
            "say": "Hello WebServer!",
        }
    }
    return errors.New("something error")
}

func main() {
    t := WebServer.Classic()
    t.Get("/", new(Action))
    t.Run()
}
```

然后在浏览器访问`http://localhost:8000`, 将会得到一个json返回
```
{"say":"Hello WebServer!"}
```

如果将上述例子中的 `true` 改为 `false`, 将会得到一个json返回
```
{"err":"something error"}
```

这段代码因为拥有一个内嵌的`WebServer.Json`，所以返回值会被自动的转成Json

## 文档

- [Manual](http://gobook.io/read/github.com/go-WebServer/manual-en-US/), And you are welcome to contribue for the book by git PR to [github.com/go-WebServer/manual-en-US](https://github.com/go-WebServer/manual-en-US)
- [操作手册](http://gobook.io/read/github.com/go-WebServer/manual-zh-CN/)，您也可以访问 [github.com/go-WebServer/manual-zh-CN](https://github.com/go-WebServer/manual-zh-CN)为本手册进行贡献
- [API Reference](https://gowalker.org/github.com/lunny/WebServer)

## 交流讨论

- QQ群：369240307
- [论坛](https://groups.google.com/forum/#!forum/go-WebServer)

## 使用案例
- [Wego](https://github.com/go-WebServer/wego)  WebServer结合[xorm](http://www.xorm.io/)开发的论坛
- [Pugo](https://github.com/go-xiaohei/pugo) 博客
- [DBWeb](https://github.com/go-xorm/dbweb) 基于Web的数据库管理工具
- [Godaily](http://godaily.org) - [github](https://github.com/godaily/news) RSS聚合工具
- [Gos](https://github.com/go-WebServer/gos)  简易的Web静态文件服务端
- [GoFtpd](https://github.com/goftp/ftpd) - 纯Go的跨平台FTP服务器

## 中间件列表

[中间件](https://github.com/WebServer-contrib)可以重用代码并且简化工作：

- [recovery](https://github.com/lunny/WebServer/wiki/Recovery) - recover after panic
- [compress](https://github.com/lunny/WebServer/wiki/Compress) - Gzip & Deflate compression
- [static](https://github.com/lunny/WebServer/wiki/Static) - Serves static files
- [logger](https://github.com/lunny/WebServer/wiki/Logger) - Log the request & inject Logger to action struct
- [param](https://github.com/lunny/WebServer/wiki/Params) - get the router parameters
- [return](https://github.com/lunny/WebServer/wiki/Return) - Handle the returned value smartlly
- [context](https://github.com/lunny/WebServer/wiki/Context) - Inject context to action struct
- [session](https://github.com/WebServer-contrib/session) - [![Build Status](https://drone.io/github.com/WebServer-contrib/session/status.png)](https://drone.io/github.com/WebServer-contrib/session/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/session)](http://gocover.io/github.com/WebServer-contrib/session) Session manager, [session-redis](http://github.com/WebServer-contrib/session-redis), [session-nodb](http://github.com/WebServer-contrib/session-nodb), [session-ledis](http://github.com/WebServer-contrib/session-ledis), [session-ssdb](http://github.com/WebServer-contrib/session-ssdb)
- [xsrf](https://github.com/WebServer-contrib/xsrf) - [![Build Status](https://drone.io/github.com/WebServer-contrib/xsrf/status.png)](https://drone.io/github.com/WebServer-contrib/xsrf/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/xsrf)](http://gocover.io/github.com/WebServer-contrib/xsrf) Generates and validates csrf tokens
- [binding](https://github.com/WebServer-contrib/binding) - [![Build Status](https://drone.io/github.com/WebServer-contrib/binding/status.png)](https://drone.io/github.com/WebServer-contrib/binding/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/binding)](http://gocover.io/github.com/WebServer-contrib/binding) Bind and validates forms
- [renders](https://github.com/WebServer-contrib/renders) - [![Build Status](https://drone.io/github.com/WebServer-contrib/renders/status.png)](https://drone.io/github.com/WebServer-contrib/renders/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/renders)](http://gocover.io/github.com/WebServer-contrib/renders) Go template engine
- [dispatch](https://github.com/WebServer-contrib/dispatch) - [![Build Status](https://drone.io/github.com/WebServer-contrib/dispatch/status.png)](https://drone.io/github.com/WebServer-contrib/dispatch/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/dispatch)](http://gocover.io/github.com/WebServer-contrib/dispatch) Multiple Application support on one server
- [tpongo2](https://github.com/WebServer-contrib/tpongo2) - [![Build Status](https://drone.io/github.com/WebServer-contrib/tpongo2/status.png)](https://drone.io/github.com/WebServer-contrib/tpongo2/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/tpongo2)](http://gocover.io/github.com/WebServer-contrib/tpongo2) [Pongo2](https://github.com/flosch/pongo2) teamplte engine support
- [captcha](https://github.com/WebServer-contrib/captcha) - [![Build Status](https://drone.io/github.com/WebServer-contrib/captcha/status.png)](https://drone.io/github.com/WebServer-contrib/captcha/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/captcha)](http://gocover.io/github.com/WebServer-contrib/captcha) Captcha
- [events](https://github.com/WebServer-contrib/events) - [![Build Status](https://drone.io/github.com/WebServer-contrib/events/status.png)](https://drone.io/github.com/WebServer-contrib/events/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/events)](http://gocover.io/github.com/WebServer-contrib/events) Before and After
- [flash](https://github.com/WebServer-contrib/flash) - [![Build Status](https://drone.io/github.com/WebServer-contrib/flash/status.png)](https://drone.io/github.com/WebServer-contrib/flash/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/flash)](http://gocover.io/github.com/WebServer-contrib/flash) Share data between requests
- [debug](https://github.com/WebServer-contrib/debug) - [![Build Status](https://drone.io/github.com/WebServer-contrib/debug/status.png)](https://drone.io/github.com/WebServer-contrib/debug/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/debug)](http://gocover.io/github.com/WebServer-contrib/debug) show detail debug infomaton on log
- [basicauth](https://github.com/WebServer-contrib/basicauth) - [![Build Status](https://drone.io/github.com/WebServer-contrib/basicauth/status.png)](https://drone.io/github.com/WebServer-contrib/basicauth/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/basicauth)](http://gocover.io/github.com/WebServer-contrib/basicauth) basicauth middleware
- [cache](https://github.com/WebServer-contrib/cache) - [![Build Status](https://drone.io/github.com/WebServer-contrib/cache/status.png)](https://drone.io/github.com/WebServer-contrib/cache/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/cache)](http://gocover.io/github.com/WebServer-contrib/cache) cache middleware - cache-memory, cache-file, [cache-ledis](https://github.com/WebServer-contrib/cache-ledis), [cache-nodb](https://github.com/WebServer-contrib/cache-nodb), [cache-mysql](https://github.com/WebServer-contrib/cache-mysql), [cache-postgres](https://github.com/WebServer-contrib/cache-postgres), [cache-memcache](https://github.com/WebServer-contrib/cache-memcache), [cache-redis](https://github.com/WebServer-contrib/cache-redis)
- [rbac](https://github.com/WebServer-contrib/rbac) - [![Build Status](https://drone.io/github.com/WebServer-contrib/rbac/status.png)](https://drone.io/github.com/WebServer-contrib/rbac/latest) [![](http://gocover.io/_badge/github.com/WebServer-contrib/debug)](http://gocover.io/github.com/WebServer-contrib/rbac) rbac control

## License
This project is under BSD License. See the [LICENSE](LICENSE) file for the full license text.
