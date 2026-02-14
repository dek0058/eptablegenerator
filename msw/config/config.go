package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type MswConfig struct {
	CoreVersion string `json:"CoreVersion"`
}

type Config struct {
	MswConfig string `yaml:"msw_config"`
	SourceDir string `yaml:"source_dir"`
	DestDir   string `yaml:"dest_dir"`
	CsvDir    string `yaml:"csv_dir"`
}

func NewConfig() *Config {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exeDir := filepath.Dir(path)
	return &Config{
		SourceDir: exeDir,
		DestDir:   exeDir,
		CsvDir:    exeDir,
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

	if !isFileExists(config.MswConfig) {
		err = errors.Join(err, errors.New("msw_config not exists"))
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

func isFileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return err == nil && !info.IsDir()
}
