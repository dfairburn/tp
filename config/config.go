package config

import (
	"fmt"
  "os"

  "gopkg.in/yaml.v3"
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

func LoadOrNewConfig(path *string) {
	if path == nil {
		// look for config in default place
		//f, err := os.Open("/tmp/dat")
		for _, path := range PathsToCheck {
			fmt.Println(path)
		}
	}
}

func LoadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	defer f.Close()

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
