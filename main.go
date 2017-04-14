package main

import "runtime"

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	hydra := NewHydra()
	hydra.Install()
	err := hydra.Start()
	if err != nil {
		hydra.Fatal(err)
	}
	hydra.Close()
}
