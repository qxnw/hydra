package main

import (
	"fmt"
	_ "github.com/qxnw/hydra/client/rpc"
	_"github.com/qxnw/hydra/engines"
	_ "github.com/qxnw/hydra/servers/http"
	_ "github.com/qxnw/hydra/servers/cron"
	_ "github.com/qxnw/hydra/servers/mqc"
	_ "github.com/qxnw/hydra/servers/rpc"
	_ "github.com/qxnw/hydra/micro"

)



func main(){
	fmt.Println("abc")
}