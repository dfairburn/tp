package cmd

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
		"The of the name of the variable, followed by a semicolon, followed by the override value, " +
		"e.g:\n" +
		"--overrides=variableName1:overriddenValue1,variableName2:overriddenValue2"
	useCmd = &cobra.Command{
		Use:   "use",
		Short: "Uses a given template to send a curl request",
		Long:  `Uses a given or chosen template, interpolates the variables and sends a curl request`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("no path provided")
			}
			var vars map[interface{}]interface{}

			if c.VariableDefinitionFile == "" {
				// if variableDefinitionFile isn't set in config, don't add it to paths to be checked
				vars = config.LoadVars(logger, varsFile)
			} else {
				vars = config.LoadVars(logger, varsFile, c.VariableDefinitionFile)
			}
			overrides := validateOverrides(overrides)
			logger.Info("overrides: ", overrides)

			return handlers.Use(logger, args[0], vars)
		},
	}
)

func init() {
	useCmd.Flags().StringArrayP("overrides", "o", overrides, overrideUsage)

	rootCmd.AddCommand(useCmd)
}

func validateOverrides(o []string) []string {
	return o
}
