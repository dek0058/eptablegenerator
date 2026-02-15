package main

import (
	"eptablegenerator/mswgoproject/config"
	"eptablegenerator/mswgoproject/gen"
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
		defaultPath := os.ExpandEnv(filepath.Join(filepath.Dir(ex), "config.yml"))
		if _, err := os.Stat(defaultPath); err == nil {
			c = *config.LoadConfig(defaultPath)
		} else {
			c = *config.NewConfig()
		}
	}

	if err := os.MkdirAll(c.DestDir, os.ModePerm); err != nil {
		println("Error creating destination directory:", err)
		panic(err)
	}

	if err := os.MkdirAll(c.CsvDir, os.ModePerm); err != nil {
		println("Error creating CSV directory:", err)
		panic(err)
	}

	if err := gen.Generate(&c); err != nil {
		println("Error generating MSWG project:", err)
		panic(err)
	}
}
