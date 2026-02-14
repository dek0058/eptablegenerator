package main

import (
	"eptablegenerator/ueproject/config"
	"eptablegenerator/ueproject/gen"
	"os"
)

func main() {
	var c config.Config

	if len(os.Args) > 1 {
		c = *config.LoadConfig(os.Args[1])
	} else {
		c = *config.NewConfig()
	}

	if err := gen.Generate(&c); err != nil {
		println("Error generating UE project:", err)
		panic(err)
	}
}
