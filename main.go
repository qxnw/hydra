package main

func main() {
	hydra := NewHydra()
	hydra.Install()
	err := hydra.Start()
	if err != nil {
		hydra.Fatal(err)
	}
}
