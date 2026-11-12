package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	logging "github.com/sirupsen/logrus"
)

func Env(logger *logging.Logger, path string) error {
	ed := os.Getenv(editor)
	if ed == "" {
		ed = "vim"
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating environment file directory: %w", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte("---\n"), 0644); err != nil {
			return fmt.Errorf("error creating variable file (%s): %w", path, err)
		}
	} else if err != nil {
		return fmt.Errorf("checking environment file (%s): %w", path, err)
	}

	return runEditor(ed, path)
}
