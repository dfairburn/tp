package handlers

import (
	"errors"
	"fmt"
	"github.com/dfairburn/tp/config"
	logging "github.com/sirupsen/logrus"
	"os"
)

func Env(logger *logging.Logger, c config.Config, envFile string) error {
	//varsFile, _ := config.LoadEnvironment(logger, varsFile, c.VariableDefinitionFile)
	// We should check if there are any .yaml/.yml files within the known file locations
	// If there isn't, we should create one in the default location of: ~/.tp/vars.yml
	path := c.EnvironmentFile
	if envFile != "" {
		path = envFile
	}

	ed := os.Getenv(editor)
	if ed == "" {
		ed = "vim"
	}

	if _, exists := os.Stat(path); errors.Is(exists, os.ErrNotExist) {
		err := os.WriteFile(path, []byte("---"), 0644)
		if err != nil {
			return fmt.Errorf("error creating variable file (%s): %v", path, err)
		}
	}

	return runEditor(ed, path)
}
