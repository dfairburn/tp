package handlers

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dfairburn/tp/static"
	logging "github.com/sirupsen/logrus"

	tppaths "github.com/dfairburn/tp/paths"
)

const (
	editor = "EDITOR"
)

func Open(logger *logging.Logger, templateDir, template string) error {
	expanded := tppaths.Expand(templateDir)
	d, err := os.Stat(expanded)
	if err != nil {
		return err
	}

	if !d.IsDir() {
		return fmt.Errorf("configured templates dir %s is not a directory", expanded)
	}

	var path = template
	if !filepath.IsAbs(template) {
		path = filepath.Join(expanded, template)
	}

	dir, _ := filepath.Split(path)
	err = os.MkdirAll(dir, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	if filepath.Ext(path) == ".yaml" {
		path = strings.TrimSuffix(path, ".yaml") + ".yml"
	}

	if filepath.Ext(path) == "" {
		path = path + ".yml"
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
