WebServer [简体中文](README_CN.md)
=======================

[![CircleCI](https://circleci.com/gh/lunny/WebServer/tree/master.svg?style=svg)](https://circleci.com/gh/lunny/WebServer/tree/master)  [![](http://gocover.io/_badge/github.com/lunny/WebServer)](http://gocover.io/github.com/lunny/WebServer) [![Join the chat at https://gitter.im/lunny/WebServer](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/lunny/WebServer?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

![WebServer Logo](logo.png)

package server is a micro & pluggable web framework for Go.

##### Current version: v0.5.0   [Version History](https://github.com/lunny/WebServer/releases)

## Getting Started

To install WebServer:

    go get github.com/lunny/WebServer

A classic usage of WebServer below:

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

Then visit `http://localhost:8000` on your browser. You will get
```
{"say":"Hello WebServer!"}
```

If you change `true` after `if` to `false`, then you will get
```
{"err":"something error"}
```

This code will automatically convert returned map or error to a json because we has an embedded struct `WebServer.Json`.

## Features

- Powerful routing & Flexible routes combinations.
- Directly integrate with existing services.
- Easy to plugin features with modular design.
- High performance dependency injection embedded.

## Middlewares

Middlewares allow you easily plugin features for your WebServer applications.

There are already many [middlewares](https://github.com/WebServer-contrib) to simplify your work:

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

## Documentation

- [Manual](http://gobook.io/read/github.com/go-WebServer/manual-en-US/), And you are welcome to contribue for the book by git PR to [github.com/go-WebServer/manual-en-US](https://github.com/go-WebServer/manual-en-US)
- [操作手册](http://gobook.io/read/github.com/go-WebServer/manual-zh-CN/)，您也可以访问 [github.com/go-WebServer/manual-zh-CN](https://github.com/go-WebServer/manual-zh-CN)为本手册进行贡献
- [API Reference](https://gowalker.org/github.com/lunny/WebServer)

## Discuss

- [Google Group - English](https://groups.google.com/forum/#!forum/go-WebServer)
- QQ Group - 简体中文 #369240307

## Cases

- [Wego](https://github.com/go-WebServer/wego) - Discuss Forum
- [dbweb](https://github.com/go-xorm/dbweb) - DB management web UI
- [Godaily](http://godaily.org) - [github](https://github.com/godaily/news)
- [Pugo](https://github.com/go-xiaohei/pugo) - A pugo blog
- [Gos](https://github.com/go-WebServer/gos) - Static web server
- [GoFtpd](https://github.com/goftp/ftpd) - Pure Go cross-platform ftp server

## License

This project is under BSD License. See the [LICENSE](LICENSE) file for the full license text.
