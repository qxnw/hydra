package main

import "runtime"

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	hydra := NewHydra()
	hydra.Install()
	defer hydra.Close()
	hydra.Start()

}
