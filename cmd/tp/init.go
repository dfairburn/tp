package main

import (
	"os"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/static"
	"github.com/spf13/cobra"
)

var (
	// local flags
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Init sets up the default directory structure and config",
		Long:  "Init sets up the default directory structure and config",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := os.MkdirAll(config.DefaultDirectory, os.ModePerm)
			if err != nil {
				return err
			}

			logger.SetOutput(os.Stdout)
			logger.
				WithField("path", config.DefaultDirectory).
				Info("initialized default directory")

			if _, err := os.Stat(config.DefaultConfigFile); err != nil && os.IsNotExist(err) {
				err = os.WriteFile(config.DefaultConfigFile, static.DefaultConfig, os.ModePerm)
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			}

			logger.
				WithField("path", config.DefaultConfigFile).
				Info("initialized default config file")

			err = os.MkdirAll(config.DefaultTemplatesDirectory, os.ModePerm)
			if err != nil {
				return err
			}

			logger.
				WithField("path", config.DefaultTemplatesDirectory).
				Info("initialized default templates directory")

			if _, err := os.Stat(config.DefaultVarFile); err != nil && os.IsNotExist(err) {
				f, err := os.Create(config.DefaultVarFile)
				defer f.Close()
				if err != nil {
					return err
				}
				logger.
					WithField("path", config.DefaultTemplatesDirectory).
					Info("initialized default variables file")
			} else if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)
}
