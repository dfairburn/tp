package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dfairburn/tp/config"
	logging "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// global flags
	c        config.Config
	logger   = logging.New()
	cfgFile  string
	logFile  string
	varsFile string
	debug    bool
	rootCmd  = &cobra.Command{
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.ConfigHomeLocFile, "yaml file containing config")
	rootCmd.PersistentFlags().StringVar(&varsFile, "vars", config.VarHomeLocFile, "yaml file containing variable definitions")
	rootCmd.PersistentFlags().StringVar(&logFile, "log", "~/.tp/tp.log", "destination of log file")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "whether to print debug to stdout")
}

func initConfig() {
	cfg, err := config.LoadOrDefaultConfig(logger, cfgFile)
	if err != nil {
		logger.WithField("errors", err).Error("could not find any config files")
	}

	c = cfg
}

func initLogger() {
	lf := openOrCreateFile(logFile)
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

func openOrCreateFile(p string) *os.File {
	// check if we can open the file for writing straight away
	expandedPath := expandTilde(p)
	lf, err := os.Create(expandedPath)
	if err != nil {
		// create dir so we can open the file for writing
		dir := path.Dir(p)
		expandedDir := expandTilde(dir)

		err := os.Mkdir(expandedDir, 0750)
		if err != nil {
			panic(fmt.Errorf("could not create log file path: %v", err))
		}

		lf, err = os.Create(expandedPath)
		if err != nil {
			panic(fmt.Errorf("could not open log file path: %v", err))
		}
	}

	return lf
}

func expandTilde(p string) string {
	if strings.HasPrefix(p, "~/") {
		replaced := filepath.Join("$HOME", p[2:])
		return os.ExpandEnv(replaced)
	}
	return p
}
