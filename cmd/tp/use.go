package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/handlers"
	"github.com/dfairburn/tp/paths"
	"github.com/dfairburn/tp/static"
	"github.com/spf13/cobra"
)

var (
	// local flags
	overrideFlagName = "overrides"
	// TODO: --help show all available overrides
	overrides       []string
	loadedOverrides []string
	rawOutput       = false
	rawUsage        = "Outputs the raw result of the request without formatting.\n"
	overrideUsage   = "A comma separated list of variable overrides.\n" +
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
				if !filepath.IsAbs(args[0]) {
					path, err := paths.NewAbsoluteFromRelative(template, c.TemplatesDirectoryPath)
					if err != nil {
						return err
					}
					template = path
				}
			}

			if len(args) < 1 {
				re, err := regexp.Compile(static.YamlRegex)
				if err != nil {
					return err
				}

				var templates []FilePath
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

					filepath, err := NewFilePath(p, c.TemplatesDirectoryPath)
					if err != nil {
						logger.Errorf("got error from new filepath (%s): %v", p, err)
						return nil
					}

					templates = append(templates, filepath)
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
			_, vars = config.LoadEnvironment(logger, envFile, c.EnvironmentFile)

			o, err := config.ValidateOverrides(overrides, logger)
			if err != nil {
				logger.Fatalf("cannot use overrides due to errors: %v", err)
			}

			var usageErr handlers.UsageError
			err = handlers.Use(logger, template, vars, o, rawOutput)
			if err != nil && errors.As(err, &usageErr) {
				fmt.Println()
				fmt.Println(err)
				fmt.Println()
				printTemplateHelp(cmd, args)
				return nil
			}
			return err
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			var templates []string
			re, err := regexp.Compile(static.YamlRegex)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

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

				expanded := paths.Expand(c.TemplatesDirectoryPath)
				filename := stripTemplateDir(fmt.Sprintf("%s/", expanded), p)
				if filename == "" {
					return nil
				}

				templates = append(templates, filename)
				return nil
			}

			err = config.LoadTemplateFiles(logger, c.TemplatesDirectoryPath, walkFunc)
			if err != nil {
				logger.Fatalf("cannot find templates in templates dir %v, error: %v", c.TemplatesDirectoryPath, err)
				return nil, cobra.ShellCompDirectiveError
			}

			return templates, cobra.ShellCompDirectiveNoFileComp
		},
	}
)

type FilePath struct {
	Absolute string
	Relative string
}

func NewFilePath(path string, tmpDir string) (FilePath, error) {
	expanded := paths.Expand(tmpDir)
	re := regexp.MustCompile(fmt.Sprintf("(?:%s/)(?P<filename>[a-zA-Z\\/\\_]+)(?:.yml|.yaml)", expanded))
	matches := re.FindStringSubmatch(path)
	keys := re.SubexpNames()
	m := make(map[string]string)
	if len(keys) != len(matches) {
		return FilePath{}, fmt.Errorf("file not found in %s directory", expanded)
	}

	for i, key := range keys {
		m[key] = matches[i]
	}

	return FilePath{
		Absolute: path,
		Relative: m["filename"],
	}, nil
}

func init() {
	useCmd.Flags().StringSliceVarP(&overrides, overrideFlagName, "o", []string{}, overrideUsage)
	useCmd.Flags().BoolVar(&rawOutput, "raw", false, rawUsage)

	err := useCmd.RegisterFlagCompletionFunc(overrideFlagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, varMap := config.LoadEnvironment(logger, envFile, c.EnvironmentFile)
		var vars []string
		for key, _ := range varMap {
			s := key.(string)
			vars = append(vars, s)
		}

		return vars, cobra.ShellCompDirectiveDefault
	})
	if err != nil {
		logger.WithError(err).Panic("cannot provide flag completions")
	}

	useCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// cobra doesn't run these on help runs :(
		initLogger()
		initConfig()
		initEnv()

		printTemplateHelp(cmd, args)
	})
	rootCmd.AddCommand(useCmd)
}

func printTemplateHelp(cmd *cobra.Command, args []string) {
	filteredArgs := slices.DeleteFunc(args, func(s string) bool {
		return s == "use"
	})

	path, err := paths.NewAbsoluteFromRelative(filteredArgs[0], c.TemplatesDirectoryPath)
	if err != nil && !os.IsNotExist(err) {
		logger.WithError(err).Panic("generating template path")
	}

	_, err = os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		// this template does not exist, show standard usage
		fmt.Println(cmd.UsageString())
		os.Exit(0)
	} else if err != nil {
		logger.WithError(err).Panic("checking template file")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		logger.WithError(err).Panic("reading template file")
	}

	tmplUsageBlock, err := handlers.GenerateTemplateUsage(content)
	if err != nil {
		logger.WithError(err).Panic("failure generating template usage")
	}

	fmt.Println(cmd.Short)
	fmt.Println("")
	fmt.Println("Template Parameters:")
	fmt.Println(tmplUsageBlock)
	fmt.Println(cmd.UsageString())
}
