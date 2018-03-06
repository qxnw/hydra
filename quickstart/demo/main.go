package main

import (
	"github.com/qxnw/hydra/hydra"
)

func main() {
	hydra := hydra.New(loader())
	defer hydra.Close()
	hydra.Start()
}
