package test

import (
	"eptablegenerator/msw/config"
	"eptablegenerator/msw/gen"
	"os"
	"path"
	"testing"
)

func TestGeneratorMSW(t *testing.T) {
	t.Log("TestGeneratorMSW")

	p, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	c := &config.Config{
		MswConfig: path.Join(p, "WorldConfig.config"),
		SourceDir: p,
		DestDir:   p,
		CsvDir:    p,
	}

	if err := gen.Generate(c); err != nil {
		t.Fatal(err)
	}
}
