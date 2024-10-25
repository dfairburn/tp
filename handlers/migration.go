package handlers

import (
	"fmt"
	"github.com/dfairburn/tp/internal/insomnia"
	"os"
)

func HandleInsomniaMigration(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	i, err := insomnia.ParseInsomniaExport(file)
	if err != nil {
		return err
	}
	fmt.Printf("got export file: %+v\n", i)

	return nil
}
