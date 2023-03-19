package config

import (
	"errors"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	InDir    = "config/"
	HomeLoc  = "~/.tp/"
	RelLoc   = "./tp/"
	Filename = "config.yml"

	HomeLocInDirFile = HomeLoc + InDir + Filename
	HomeLocFile      = HomeLoc + Filename
	RelLocInDirFile  = RelLoc + InDir + Filename
	RelLocFile       = RelLoc + Filename
)

var (
	PathsToCheck = []string{
		HomeLocInDirFile,
		HomeLocFile,
		RelLocInDirFile,
		RelLocFile,
	}
)

func LoadOrDefaultConfig(path string) (Config, error) {
	var errs []error
	// Check path given via cli flag
	if path != "" {
		config, err := loadConfig(path)
		if err != nil {
			errs = append(errs, err)
		} else {
			return config, nil
		}
	}

	// Check default paths
	for _, p := range PathsToCheck {
		config, err := loadConfig(p)
		if err != nil {
			errs = append(errs, err)
		} else {
			return config, nil
		}
	}

	return Config{}, errors.Join(errs...)
}

func loadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

type Config struct {
	// File defining defaults for variables
	VariableDefinitionFile string `yaml:"variableDefinitionFile"`
}
