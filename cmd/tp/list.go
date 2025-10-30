package main

import (
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/static"
	"github.com/xlab/treeprint"

	"github.com/spf13/cobra"
)

var (
	// local flags
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List returns all the templates in the configured templates directory",
		Long:  "List returns all the templates in the configured templates directory",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			re, err := regexp.Compile(static.YamlRegex)
			if err != nil {
				return err
			}

			tree := treeprint.New()
			dirNodes := map[string]treeprint.Tree{}

			walkFunc := func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					if p == c.TemplatesDirectoryPath {
						// don't treat root as a node on the tree
						return nil
					}
					if parentNode, ok := dirNodes[path.Dir(p)]; ok {
						// subdirectory of existing node
						dirNodes[p] = parentNode.AddBranch(info.Name())
					} else {
						// must be subdirectory of root
						dirNodes[p] = tree.AddBranch(info.Name())
					}
					return nil
				}
				if ok := re.Match([]byte(p)); !ok {
					logger.Errorf("path %v does not have yaml extension\n", p)
					return nil
				}

				tplName := strings.TrimSuffix(info.Name(), ".yaml")
				tplName = strings.TrimSuffix(tplName, ".yml")
				dirPath := path.Dir(p)
				if dirPath == c.TemplatesDirectoryPath {
					// add loose files to root of tree
					tree.AddNode(tplName)
				} else {
					// use the appropriate tree node for subdirectories
					dirNodes[dirPath].AddNode(tplName)
				}
				return nil
			}
			err = config.LoadTemplateFiles(logger, c.TemplatesDirectoryPath, walkFunc)
			if err != nil {
				logger.Fatalf("cannot find templates in templates dir %v, error: %v", c.TemplatesDirectoryPath, err)
			}

			_, err = io.WriteString(os.Stdout, tree.String())
			return err
		},
	}
)

func init() {
	rootCmd.AddCommand(listCmd)
}
