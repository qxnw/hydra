package main

import (
	"github.com/qxnw/hydra/micro"
)

func main() {
	app := micro.NewApp(micro.WithDebug())
	app.Start()
}
