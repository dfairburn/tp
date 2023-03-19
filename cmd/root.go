package cmd

import (
	"errors"
	"fmt"
	"github.com/dfairburn/tp/config"
	logging "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"time"
)

var (
	// global flags
	c       config.Config
	logger  = logging.New()
	cfgFile string
	logFile string
	rootCmd = &cobra.Command{
		Use:   "tp",
		Short: "tp is a configurable api client",
		Long:  `some text about tp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("no subcommand given")
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initLogger, initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/tp/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log", ".tp.log", "config file (default is $HOME/tp/config.yaml)")

	//cobra.OnInitialize(initLogger, initConfig)
}

func initConfig() {
	cfg, err := config.LoadOrDefaultConfig(cfgFile)
	if err != nil {
		logger.WithField("errors", err).Error("could not find any config files")
	}

	c = cfg
}

func initLogger() {
	lf, err := os.Create(logFile)
	if err != nil {
		panic(fmt.Errorf("could not create log file: %v", err))
	}

	mw := io.MultiWriter(os.Stdout, lf)
	logger.Out = mw

	formatter := logging.JSONFormatter{
		TimestampFormat: time.RFC3339,
	}
	logger.Formatter = &formatter
}
