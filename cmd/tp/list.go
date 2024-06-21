package main

import (
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/dfairburn/tp/config"

	"github.com/spf13/cobra"
)

var (
	// local flags
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List returns all the templates in the configured templates directory",
		Long:  "List returns all the templates in the configured templates directory",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			re, err := regexp.Compile(".*\\.tmpl$")
			if err != nil {
				return err
			}

			var templates []string
			walkFunc := func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if ok := re.Match([]byte(p)); !ok {
					logger.Errorf("path %v does not have .tmpl extension\n", p)
					return nil
				}
				if info.IsDir() {
					logger.Errorf("path %v is a directory\n", p)
					return nil
				}

				templates = append(templates, p)
				return nil
			}
			err = config.LoadTemplateFiles(logger, c.TemplatesDirectoryPath, walkFunc)
			if err != nil {
				logger.Fatalf("cannot find templates in templates dir %v, error: %v", c.TemplatesDirectoryPath, err)
			}

			_, err = io.WriteString(os.Stdout, strings.Join(templates, "\n"))
			_, _ = io.WriteString(os.Stdout, "\n")
			return err
		},
	}
)

func init() {
	rootCmd.AddCommand(listCmd)
}
