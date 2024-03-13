package handlers

import (
	"errors"
	"os"

	logging "github.com/sirupsen/logrus"
)

func New(logger *logging.Logger, template string) error {
	if _, err := os.Stat(template); errors.Is(err, os.ErrNotExist) {
		_, err := os.Create(template)
		if err != nil {
			return err
		}

		//f.WriteString()

		// path/to/whatever does not exist
	}
	return nil
}
