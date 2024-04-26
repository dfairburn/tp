package handlers

import (
	"errors"
	"fmt"
	"github.com/dfairburn/tp/config"
	tppaths "github.com/dfairburn/tp/paths"
	logging "github.com/sirupsen/logrus"
	"os"
)

func Vars(logger *logging.Logger, c config.Config, varsFile string) error {
	path, _ := config.LoadVars(logger, varsFile, c.VariableDefinitionFile)
	expanded := tppaths.Expand(path)

	if _, err := os.Stat(expanded); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error finding config file %s: %v", expanded, err)
	}
	ed := os.Getenv(editor)
	if ed == "" {
		ed = "vim"
	}

	return runEditor(ed, expanded)
}
