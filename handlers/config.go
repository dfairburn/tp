package handlers

import (
	"errors"
	"fmt"
	"os"

	tppaths "github.com/dfairburn/tp/paths"
	logging "github.com/sirupsen/logrus"
)

func Config(logger *logging.Logger, path string) error {
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
