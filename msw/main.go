package main

import (
	"eptablegenerator/msw/config"
	"eptablegenerator/msw/gen"
	"os"
	"path/filepath"
)

func main() {
	var c config.Config

	if len(os.Args) > 1 {
		c = *config.LoadConfig(os.Args[1])
	} else {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		defaultPath := filepath.Join(filepath.Dir(ex), "config.yml")
		if _, err := os.Stat(defaultPath); err == nil {
			c = *config.LoadConfig(defaultPath)
		} else {
			c = *config.NewConfig()
		}
	}

	if err := gen.Generate(&c); err != nil {
		println("Error generating MSW project:", err)
		panic(err)
	}
}
