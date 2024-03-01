package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

func LoadTemplateFiles(l *logging.Logger, dirPath string) ([]string, error) {
	var templates []string
	path := tppaths.Expand(dirPath)
	re, err := regexp.Compile(".*\\.tmpl$")
	if err != nil {
		return nil, err
	}
	err = filepath.Walk(path,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if ok := re.Match([]byte(p)); !ok {
				l.Errorf("path %v does not have .tmpl extension\n", path)
				return nil
			}
			if info.IsDir() {
				l.Errorf("path %v is a directory\n", path)
				return nil
			}

			templates = append(templates, p)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if len(templates) < 1 {
		return nil, fmt.Errorf("no templates in template dir %v", dirPath)
	}

	return templates, nil
}
