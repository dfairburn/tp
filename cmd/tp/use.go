package main

import (
	"fmt"
	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/handlers"
	"github.com/dfairburn/tp/paths"
	"github.com/spf13/cobra"
	"os"
	"regexp"
)

var (
	// local flags
	overrideFlagName = "overrides"
	// TODO: --help show all available overrides
	overrides       []string
	loadedOverrides []string
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
				if isAbsolutePath(args[0]) {
					template = args[0]
				}

				absolute := NewAbsoluteFromRelative(c.TemplatesDirectoryPath, args[0])
				template = absolute
			}

			if len(args) < 1 {
				re, err := regexp.Compile(".*\\.tmpl$")
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
			_, vars = config.LoadVars(logger, varsFile, c.VariableDefinitionFile)
			o, err := config.ValidateOverrides(overrides, logger)
			if err != nil {
				logger.Fatalf("cannot use overrides due to errors: %v", err)
			}

			return handlers.Use(logger, template, vars, o)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			var templates []string
			re, err := regexp.Compile(".*\\.tmpl$")
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

				fmt.Printf("%v\n", filename)
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
	re := regexp.MustCompile(fmt.Sprintf("(?:%s/)(?P<filename>[a-zA-Z\\/\\_]+)(?:.tmpl)", expanded))
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

func NewAbsoluteFromRelative(tmpDir string, p string) string {
	re := regexp.MustCompile(".*\\.tmpl$")
	if re.Match([]byte(p)) {
		return fmt.Sprintf("%s/%s", tmpDir, p)
	} else {
		return fmt.Sprintf("%s/%s.tmpl", tmpDir, p)
	}
}

func isAbsolutePath(path string) bool {
	re := regexp.MustCompile("^\\/")
	return re.Match([]byte(path))
}

func init() {
	useCmd.Flags().StringSliceVarP(&overrides, overrideFlagName, "o", []string{}, overrideUsage)
	err := useCmd.RegisterFlagCompletionFunc(overrideFlagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, varMap := config.LoadVars(logger, varsFile, c.VariableDefinitionFile)
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

	rootCmd.AddCommand(useCmd)
}
