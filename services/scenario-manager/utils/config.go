package utils

import (
	"flag"
	"fmt"
	"os"

	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	"gopkg.in/yaml.v3"
)

var cfg *entities.AppConfig

func GetConfig() *entities.AppConfig {
	return cfg
}

func NewConfig(configPath string) (*entities.AppConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a file", path)
	}
	return nil
}

func ParseFlags() (string, error) {
	var configPath string

	flag.StringVar(&configPath, "config", "./config.yaml", "path to config file")

	flag.Parse()

	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	return configPath, nil
}
