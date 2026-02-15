package test

import (
	"eptablegenerator/mswgoproject/config"
	"eptablegenerator/mswgoproject/gen"
	"os"
	"testing"
)

func TestGeneratorMSWGProject(t *testing.T) {
	t.Log("TestGeneratorMSWGProject")

	p, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	c := &config.Config{
		PackageName: "table",
		SourceDir:   p,
		DestDir:     p,
		CsvDir:      p,
	}

	if err := gen.Generate(c); err != nil {
		t.Fatal(err)
	}
}
