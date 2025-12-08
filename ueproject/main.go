package main

import (
	"eptablegenerator/table/config"
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

	if err := gen.GenerateUE(&c); err != nil {
		println("Error generating UE project:", err)
		panic(err)
	}
}
