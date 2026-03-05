package handlers

import (
	"errors"
	"github.com/dfairburn/tp/static"
	logging "github.com/sirupsen/logrus"
	"os"
	"os/exec"

	tppaths "github.com/dfairburn/tp/paths"
)

const (
	editor = "EDITOR"
)

func Open(logger *logging.Logger, templateDir, template string) error {
	path, err := tppaths.NewAbsoluteFromRelative(template, templateDir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(path, static.DefaultTemplate, 0644)
		if err != nil {
			return err
		}
	} else {
		ed := os.Getenv(editor)
		if ed == "" {
			ed = "vim"
		}

		return runEditor(ed, path)
	}

	ed := os.Getenv(editor)
	if ed == "" {
		ed = "vim"
	}

	return runEditor(ed, path)
}

func runEditor(ed string, path string) error {
	cmd := exec.Command(ed, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
