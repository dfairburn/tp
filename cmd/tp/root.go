package main

import (
	"errors"
	"github.com/ktr0731/go-fuzzyfinder"
	"io"
	"log"
	"os"
	"time"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/paths"

	logging "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// global flags
	c          config.Config
	configPath string
	varsPath   string
	logger     = logging.New()
	cfgFile    string
	logFile    string
	varsFile   string
	debug      bool
	rootCmd    = &cobra.Command{
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultConfigFile, "yaml file containing config")
	rootCmd.PersistentFlags().StringVar(&varsFile, "vars", config.DefaultVarFile, "yaml file containing variable definitions")
	rootCmd.PersistentFlags().StringVar(&logFile, "log", config.DefaultLogFile, "destination of log file")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "whether to print debug to stdout")
}

func initConfig() {
	cfg, path, err := config.LoadOrDefaultConfig(logger, cfgFile)
	if err != nil {
		logger.WithField("errors", err).Error("could not find any config files")
	}

	configPath = path
	expandedTemplateDirPath := paths.Expand(cfg.TemplatesDirectoryPath)
	expandedVariablePath := paths.Expand(cfg.VariableDefinitionFile)

	c = config.Config{
		VariableDefinitionFile: expandedVariablePath,
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
