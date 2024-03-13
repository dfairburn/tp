package main

import (
	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/handlers"
	"github.com/spf13/cobra"
	"os"
	"regexp"
)

var (
	// local flags
	// TODO: --help show all available overrides
	overrides     []string
	overrideUsage = "A comma separated list of variable overrides.\n" +
		"A variable override is a list of strings in the following format:\n" +
		"The of the name of the variable, followed by a colon, followed by the override value, e.g:\n" +
		"--overrides=variableName1:overriddenValue1,variableName2:overriddenValue2" +
		"or you can repeatedly use the flag like so:\n" +
		"-o=override1:value1 -o=override2:override2"

	useCmd = &cobra.Command{
		Use:   "use",
		Short: "Uses a given template to send a curl request",
		Long:  `Uses a given or chosen template, interpolates the variables and sends an http request`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			var template string

			if len(args) == 1 {
				template = args[0]
			}

			if len(args) < 1 {
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

				t, err := fzf(templates)
				if err != nil {
					logger.Fatalf("cannot fuzzyfind templates %v", err)
				}
				template = t
			}

			if len(args) > 1 {
				logger.Fatalf("too many arguments (%d) given to `use`, expected 1", len(args))
			}

			var vars map[interface{}]interface{}
			vars = config.LoadVars(logger, varsFile, c.VariableDefinitionFile)
			overrides, err := config.ValidateOverrides(overrides, logger)
			if err != nil {
				logger.Fatalf("cannot use overrides due to errors: %v", err)
			}

			return handlers.Use(logger, template, vars, overrides)
		},
	}
)

func init() {
	useCmd.Flags().StringSliceVarP(&overrides, "overrides", "o", []string{}, overrideUsage)

	rootCmd.AddCommand(useCmd)
}
