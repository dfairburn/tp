package main

import (
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/ktr0731/go-fuzzyfinder"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/paths"

	logging "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// global flags
	c          config.Config
	configPath string
	logger     = logging.New()
	cfgFile    string
	logFile    string
	env        string
	envFile    string
	debug      bool
	rootCmd    = &cobra.Command{
		Use:   "tp",
		Short: "Tp is a configurable api client",
		Long:  "Tp is a command-line utility to create and re-use templated http requests.\nTp was created as an alternative to GUI based API clients, to allow for easier automation and scripting of api requests.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("no subcommand given")
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initLogger, initConfig, initEnv)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", config.DefaultConfigFile, "yaml file containing config")
	rootCmd.PersistentFlags().StringVar(&envFile, "envFile", config.DefaultEnvPath, "yaml file containing variable definitions")
	rootCmd.PersistentFlags().StringVarP(&env, "env", "e", config.DefaultEnv, "a string dictating which env file to use")
	rootCmd.PersistentFlags().StringVar(&logFile, "log", config.DefaultLogFile, "destination of log file")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "whether to print debug to stdout")
}

func initEnv() {
	if env != "" {
		envFile = paths.Expand(config.HomeLoc + env + config.YmlExt)
	}
}

func initConfig() {
	cfg, path, err := config.LoadOrDefaultConfig(logger, cfgFile)
	if err != nil {
		logger.WithField("errors", err).Error("could not find any config files")
	}

	configPath = path
	expandedTemplateDirPath := paths.Expand(cfg.TemplatesDirectoryPath)
	expandedEnvironmentFilePath := paths.Expand(cfg.EnvironmentFile)

	c = config.Config{
		EnvironmentFile:        expandedEnvironmentFilePath,
		TemplatesDirectoryPath: expandedTemplateDirPath,
	}
}

func initLogger() {
	lf := paths.OpenOrCreateFile(logFile)
	var w io.Writer = lf
	if debug {
		w = io.MultiWriter(os.Stdout, lf)
	}
	logger.Out = w

	formatter := logging.JSONFormatter{
		TimestampFormat: time.RFC3339,
	}
	logger.Formatter = &formatter
}

func fzf(paths []FilePath) (string, error) {
	template, err := fuzzyfinder.Find(paths,
		func(i int) string {
			return paths[i].Relative
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	return paths[template].Absolute, err
}
