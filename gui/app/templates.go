package app

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/handlers"
	"github.com/dfairburn/tp/static"
	"gopkg.in/yaml.v3"
)

// GetTemplates returns all templates as a tree structure
func (a *App) GetTemplates() ([]*TemplateItem, error) {
	items := make([]*TemplateItem, 0)
	dirMap := make(map[string]*TemplateItem)

	re, err := regexp.Compile(static.YamlRegex)
	if err != nil {
		return nil, err
	}

	walkFunc := func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if p == a.templatesDir {
			return nil
		}

		relPath, _ := filepath.Rel(a.templatesDir, p)
		depth := strings.Count(relPath, string(os.PathSeparator))

		if info.IsDir() {
			item := &TemplateItem{
				Name:         info.Name(),
				AbsolutePath: p,
				RelativePath: relPath,
				IsDir:        true,
				Depth:        depth,
				Children:     make([]*TemplateItem, 0),
			}
			dirMap[p] = item

			parentPath := filepath.Dir(p)
			if parent, ok := dirMap[parentPath]; ok {
				parent.Children = append(parent.Children, item)
			} else {
				items = append(items, item)
			}
			return nil
		}

		if !re.MatchString(p) {
			return nil
		}

		item := &TemplateItem{
			Name:         strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())),
			AbsolutePath: p,
			RelativePath: relPath,
			IsDir:        false,
			Depth:        depth,
		}

		if content, err := os.ReadFile(p); err == nil {
			var tmpl handlers.Template
			if yaml.Unmarshal(content, &tmpl) == nil {
				item.Method = strings.ToUpper(tmpl.Method)
				if item.Method == "" {
					item.Method = "GET"
				}
				item.URL = tmpl.Url
				item.Headers = tmpl.Headers
				item.Body = tmpl.Body
				item.Descriptions = tmpl.Descriptions
			}
		}

		parentPath := filepath.Dir(p)
		if parent, ok := dirMap[parentPath]; ok {
			parent.Children = append(parent.Children, item)
		} else {
			items = append(items, item)
		}
		return nil
	}

	if err = config.LoadTemplateFiles(a.logger, a.templatesDir, walkFunc); err != nil {
		return nil, err
	}

	sortItems(items)
	return items, nil
}

func sortItems(items []*TemplateItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	for _, item := range items {
		if item.IsDir && len(item.Children) > 0 {
			sortItems(item.Children)
		}
	}
}

// GetTemplate returns details of a specific template
func (a *App) GetTemplate(absolutePath string) (*TemplateResponse, error) {
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}

	var tmpl handlers.Template
	if err := yaml.Unmarshal(content, &tmpl); err != nil {
		return nil, err
	}

	method := strings.ToUpper(tmpl.Method)
	if method == "" {
		method = "GET"
	}

	return &TemplateResponse{
		Name:         filepath.Base(absolutePath),
		AbsolutePath: absolutePath,
		Method:       method,
		URL:          tmpl.Url,
		Headers:      tmpl.Headers,
		Body:         tmpl.Body,
		Descriptions: tmpl.Descriptions,
		RawContent:   string(content),
	}, nil
}

// CreateTemplate creates a new template file with default content
func (a *App) CreateTemplate(name string, parentDir string) (string, error) {
	dir := a.templatesDir
	if parentDir != "" {
		dir = parentDir
	}

	filename := name
	if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") {
		filename = filename + ".yaml"
	}

	fullPath := filepath.Join(dir, filename)
	if _, err := os.Stat(fullPath); err == nil {
		return "", os.ErrExist
	}

	content := `# Template: ` + name + `
method: GET
url: https://api.example.com/endpoint
headers:
  Content-Type: application/json
body: |
  {}
`
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return "", err
	}
	return fullPath, nil
}

// CreateFolder creates a new folder in the templates directory
func (a *App) CreateFolder(name string, parentDir string) (string, error) {
	dir := a.templatesDir
	if parentDir != "" {
		dir = parentDir
	}

	fullPath := filepath.Join(dir, name)
	if _, err := os.Stat(fullPath); err == nil {
		return "", os.ErrExist
	}

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", err
	}
	return fullPath, nil
}

// DeleteTemplate deletes a template file or empty folder
func (a *App) DeleteTemplate(absolutePath string) error {
	if !strings.HasPrefix(absolutePath, a.templatesDir) {
		return os.ErrPermission
	}
	return os.Remove(absolutePath)
}

// RenameItem renames a template file or folder
func (a *App) RenameItem(absolutePath string, newName string) (string, error) {
	if !strings.HasPrefix(absolutePath, a.templatesDir) {
		return "", os.ErrPermission
	}

	newPath := filepath.Join(filepath.Dir(absolutePath), newName)
	if _, err := os.Stat(newPath); err == nil {
		return "", os.ErrExist
	}

	if err := os.Rename(absolutePath, newPath); err != nil {
		return "", err
	}
	return newPath, nil
}

// MoveItem moves a template or folder to a new parent directory
func (a *App) MoveItem(sourcePath string, targetDir string) (string, error) {
	if !strings.HasPrefix(sourcePath, a.templatesDir) {
		return "", os.ErrPermission
	}
	if !strings.HasPrefix(targetDir, a.templatesDir) && targetDir != a.templatesDir {
		return "", os.ErrPermission
	}

	newPath := filepath.Join(targetDir, filepath.Base(sourcePath))
	if sourcePath == newPath {
		return sourcePath, nil
	}
	if strings.HasPrefix(targetDir, sourcePath) {
		return "", fmt.Errorf("cannot move folder into itself")
	}
	if _, err := os.Stat(newPath); err == nil {
		return "", os.ErrExist
	}

	if err := os.Rename(sourcePath, newPath); err != nil {
		return "", err
	}
	return newPath, nil
}

// SaveTemplate saves template content to disk
func (a *App) SaveTemplate(absolutePath string, content string) error {
	if !strings.HasPrefix(absolutePath, a.templatesDir) {
		return os.ErrPermission
	}
	return os.WriteFile(absolutePath, []byte(content), 0644)
}

// OpenInEditor opens a file in the system default text editor
func (a *App) OpenInEditor(absolutePath string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{"-t", absolutePath}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", "", absolutePath}
	default:
		cmd = "xdg-open"
		args = []string{absolutePath}
	}
	return exec.Command(cmd, args...).Start()
}

// GetHTTPMethods returns the supported HTTP methods
func (a *App) GetHTTPMethods() []string {
	return []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
	}
}

// GetAppInfo returns application metadata
func (a *App) GetAppInfo() map[string]string {
	return map[string]string{
		"name":    "tp-gui",
		"version": "0.1.0",
		"built":   time.Now().Format(time.RFC3339),
	}
}
