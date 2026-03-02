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

func LoadOrDefaultConfig(logger *logging.Logger, paths ...string) (Config, string, error) {
	// this gives precedent to paths passed in via config and flags, then processes the default file paths
	paths = append(paths, buildConfigPaths().paths...)
	f, path, err := tryFiles(logger, paths...)
	if err != nil {
		logger.Error(err)
	}

	config, err := loadConfig(f)
	return config, path, err
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
	EnvironmentFile        string `yaml:"environmentFile"`
	TemplatesDirectoryPath string `yaml:"templatesDirectoryPath"`
}

func tryFiles(logger *logging.Logger, paths ...string) (*os.File, string, error) {
	var expandedPaths []string
	for _, p := range paths {
		expanded := tppaths.Expand(p)
		expandedPaths = append(expandedPaths, expanded)

		f, err := os.Open(expanded)
		if err != nil {
			logger.Warnf("error: %v", err)
			continue
		}

		return f, expanded, nil
	}

	return nil, "", fmt.Errorf("no files existed: %s\n", strings.Join(expandedPaths, ", "))
}

func LoadTemplateFiles(l *logging.Logger, dirPath string, walkFunc func(string, os.FileInfo, error) error) error {
	path := tppaths.Expand(dirPath)

	// Wrap the user's walkFunc to skip hidden files and directories
	wrappedWalkFunc := func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return walkFunc(p, info, err)
		}

		// Skip hidden files and directories (starting with '.')
		if len(info.Name()) > 0 && info.Name()[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		return walkFunc(p, info, err)
	}

	err := filepath.Walk(path, wrappedWalkFunc)
	if err != nil {
		return err
	}

	return nil
}
