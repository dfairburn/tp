package main

import (
	"errors"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/handlers"
	"github.com/spf13/cobra"
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
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("no path provided")
			}
			var vars map[interface{}]interface{}

			vars = config.LoadVars(logger, varsFile, c.VariableDefinitionFile)
			overrides, err := config.ValidateOverrides(overrides, logger)
			if err != nil {
				logger.Fatalf("cannot use overrides due to errors: %v", err)
			}

			return handlers.Use(logger, args[0], vars, overrides)
		},
	}
)

func init() {
	useCmd.Flags().StringSliceVarP(&overrides, "overrides", "o", []string{}, overrideUsage)

	rootCmd.AddCommand(useCmd)
}
