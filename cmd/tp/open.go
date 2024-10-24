package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/handlers"
	tppaths "github.com/dfairburn/tp/paths"
	"github.com/spf13/cobra"
)

var (
	// local flags
	newCmd = &cobra.Command{
		Use:   "open",
		Short: "Open creates a new template or opens an existing one",
		Long:  "Open creates a new template, or opens an existing one and loads it into your configured editor (configured by $EDITOR, default vim)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("need new path to be returned for new template")
			}

			filename := args[0]
			return handlers.Open(logger, c.TemplatesDirectoryPath, filename)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			var templates []string
			re, err := regexp.Compile("(?:.*\\.yaml$|.*\\.yml)")
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			walkFunc := func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if ok := re.Match([]byte(p)); !ok {
					logger.Errorf("path %v does not have yaml extension\n", p)
					return nil
				}
				if info.IsDir() {
					logger.Errorf("path %v is a directory\n", p)
					return nil
				}

				expanded := tppaths.Expand(c.TemplatesDirectoryPath)
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

func init() {
	newCmd.Flags().StringSliceVarP(&overrides, "overrides", "o", []string{}, overrideUsage)

	rootCmd.AddCommand(newCmd)
}

func stripTemplateDir(tmpDir, path string) string {
	re := regexp.MustCompile(fmt.Sprintf("(?:%s)(?P<filename>[0-9a-zA-Z\\/\\.\\_\\-]+)", tmpDir))
	matches := re.FindStringSubmatch(path)
	keys := re.SubexpNames()
	m := make(map[string]string)
	if len(keys) != len(matches) {
		return ""
	}

	for i, key := range keys {
		m[key] = matches[i]
	}

	return m["filename"]
}
