package test

import (
	"eptablegenerator/ueproject/config"
	"eptablegenerator/ueproject/gen"
	"os"
	"testing"
)

func TestGeneratorUE(t *testing.T) {
	t.Log("TestGeneratorUE")

	p, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	c := config.Config{
		ProjectName: "test",
		SourceDir:   p,
		DestDir:     p,
	}

	if err := gen.Generate(&c); err != nil {
		t.Error(err)
	}
}
