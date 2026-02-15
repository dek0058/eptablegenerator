package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PackageName string `yaml:"package_name"`
	SourceDir   string `yaml:"source_dir"`
	DestDir     string `yaml:"dest_dir"`
	CsvDir      string `yaml:"csv_dir"`
}

func NewConfig() *Config {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exeDir := filepath.Dir(path)
	return &Config{
		PackageName: "table",
		SourceDir:   exeDir,
		DestDir:     exeDir,
		CsvDir:      exeDir,
	}
}

func LoadConfig(filePath string) *Config {
	var err error

	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}

	if config.PackageName == "" {
		err = errors.Join(err, errors.New("package_name is empty"))
	}

	if !isDirExists(config.SourceDir) {
		err = errors.Join(err, errors.New("source_dir not exists"))
	}

	if !isDirExists(config.DestDir) {
		err = errors.Join(err, errors.New("dest_dir not exists"))
	}

	if !isDirExists(config.CsvDir) {
		err = errors.Join(err, errors.New("csv_dir not exists"))
	}

	if err != nil {
		panic(err)
	}

	return &config
}

func isDirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return err == nil && info.IsDir()
}
