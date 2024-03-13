package main

import (
	"errors"
	"github.com/dfairburn/tp/handlers"
	"github.com/spf13/cobra"
)

var (
	// local flags
	newCmd = &cobra.Command{
		Use:   "new",
		Short: "New creates a new template",
		Long:  "New creates a new template and opens your text editor with a default template",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("need new path to be returned for new template")
			}

			dir := args[0]

			//if len(args) < 1 {
			//	var dirs []string
			//	walkFunc := func(p string, info os.FileInfo, err error) error {
			//		if err != nil {
			//			return err
			//		}
			//		if info.IsDir() {
			//			dirs = append(dirs, p)
			//			return nil
			//		}
			//
			//		logger.Errorf("path %v isn't a directory\n", p)
			//		return nil
			//	}
			//
			//	err := config.LoadTemplateFiles(logger, c.TemplatesDirectoryPath, walkFunc)
			//	if err != nil {
			//		logger.Fatalf("cannot find templates in templates dir %v, error: %v", c.TemplatesDirectoryPath, err)
			//	}
			//
			//	d, err := fzf(dirs)
			//	if err != nil {
			//		logger.Fatalf("cannot fuzzyfind templates %v", err)
			//	}
			//	dir = d
			//}

			//var vars map[interface{}]interface{}
			//vars = config.LoadVars(logger, varsFile, c.VariableDefinitionFile)
			//overrides, err := config.ValidateOverrides(overrides, logger)
			//if err != nil {
			//	logger.Fatalf("cannot use overrides due to errors: %v", err)
			//}

			return handlers.New(logger, dir)
		},
	}
)

func init() {
	newCmd.Flags().StringSliceVarP(&overrides, "overrides", "o", []string{}, overrideUsage)

	rootCmd.AddCommand(newCmd)
}
