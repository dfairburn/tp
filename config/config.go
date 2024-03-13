package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	tppaths "github.com/dfairburn/tp/paths"

	logging "github.com/sirupsen/logrus"
)

func LoadOrDefaultConfig(logger *logging.Logger, paths ...string) (Config, error) {
	// this gives precedent to paths passed in via config and flags, then processes the default file paths
	paths = append(paths, buildConfigPaths().paths...)
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
	TemplatesDirectoryPath string `yaml:"templatesDirectoryPath"`
}

func tryFiles(logger *logging.Logger, paths ...string) (*os.File, error) {
	var expandedPaths []string
	for _, p := range paths {
		expanded := tppaths.Expand(p)
		expandedPaths = append(expandedPaths, expanded)

		f, err := os.Open(expanded)
		if err != nil {
			logger.Warnf("error: %v", err)
			continue
		}

		return f, nil
	}

	return nil, fmt.Errorf("no files existed: %s\n", strings.Join(expandedPaths, ", "))
}

func LoadTemplateFiles(l *logging.Logger, dirPath string, walkFunc func(string, os.FileInfo, error) error) error {
	path := tppaths.Expand(dirPath)
	err := filepath.Walk(path, walkFunc)
	if err != nil {
		return err
	}

	return nil
}
