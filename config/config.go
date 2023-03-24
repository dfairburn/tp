package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"

	logging "github.com/sirupsen/logrus"
)

const (
	ConfigDir      = "config/"
	ConfigFilename = "config.yml"

	LogFilename = "tp.log"

	VarDir      = "vars/"
	VarFilename = "vars.yml"

	HomeLoc = "~/.tp/"
	RelLoc  = "./tp/"

	ConfigHomeLocInDirFile = HomeLoc + ConfigDir + ConfigFilename
	ConfigHomeLocFile      = HomeLoc + ConfigFilename
	ConfigRelLocInDirFile  = RelLoc + ConfigDir + ConfigFilename
	ConfigRelLocFile       = RelLoc + ConfigFilename

	DefaultLogFile = HomeLoc + LogFilename
)

var (
	configPathsToCheck = []string{
		ConfigHomeLocInDirFile,
		ConfigHomeLocFile,
		ConfigRelLocInDirFile,
		ConfigRelLocFile,
	}
)

func LoadOrDefaultConfig(logger *logging.Logger, paths ...string) (Config, error) {
	// this gives precedent to paths passed in via config and flags, then processes the default file paths
	paths = append(paths, configPathsToCheck...)
	f, err := tryFiles(logger, paths...)
	if err != nil {
		logger.Error(err)
	}

	return loadConfig(f)
}

func loadConfig(f *os.File) (Config, error) {
	var cfg Config
	decoder := yaml.NewDecoder(f)
	err := decoder.Decode(&cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

type Config struct {
	// File defining defaults for variables
	VariableDefinitionFile string `yaml:"variableDefinitionFile"`
}

func tryFiles(logger *logging.Logger, paths ...string) (*os.File, error) {
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			logger.Warnf("error: %v", err)
			continue
		}
		return f, nil
	}

	return nil, fmt.Errorf("no files existed: %s\n", strings.Join(paths, ","))
}
