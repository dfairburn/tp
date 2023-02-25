package cmd

import (
	"errors"
	"log"

	"github.com/spf13/cobra"

	"github.com/dfairburn/tp/config"
)

var (
	// global flags
	c       config.Config
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "tp",
		Short: "tp is a configurable api client",
		Long:  `some text about tp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("No subcommand given")
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/tp/config.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			log.Fatalf("could not load given config: %v\n", err)
		}

		c = cfg
	}
}
