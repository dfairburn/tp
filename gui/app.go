package main

import (
	"context"
	"fmt"
	"io"
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
	"github.com/dfairburn/tp/paths"
	"github.com/dfairburn/tp/static"
	logging "github.com/sirupsen/logrus"
	"github.com/tidwall/pretty"
	"gopkg.in/yaml.v3"
)

// App struct holds the application state
type App struct {
	ctx          context.Context
	logger       *logging.Logger
	config       config.Config
	configPath   string
	vars         map[interface{}]interface{}
	templatesDir string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Setup logger
	a.logger = logging.New()
	a.logger.SetLevel(logging.InfoLevel)

	// Try to set up log file
	if logFile, err := os.OpenFile(paths.Expand("~/.tp/tp-gui.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		a.logger.Out = logFile
	} else {
		a.logger.Out = io.Discard
	}

	// Load configuration
	cfg, configPath, err := config.LoadOrDefaultConfig(a.logger)
	if err != nil {
		a.logger.Warnf("Error loading config: %v", err)
	}
	a.config = cfg
	a.configPath = configPath

	// Set templates directory
	a.templatesDir = paths.Expand(cfg.TemplatesDirectoryPath)
	if a.templatesDir == "" {
		a.templatesDir = paths.Expand("~/.tp/templates")
	}

	// Load environment variables
	_, a.vars = config.LoadEnvironment(a.logger, cfg.EnvironmentFile)
}

// TemplateItem represents a template for the frontend
type TemplateItem struct {
	Name         string            `json:"name"`
	AbsolutePath string            `json:"absolutePath"`
	RelativePath string            `json:"relativePath"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	Descriptions map[string]string `json:"descriptions"`
	IsDir        bool              `json:"isDir"`
	Depth        int               `json:"depth"`
	Children     []*TemplateItem   `json:"children"`
}

// TemplateResponse represents the parsed template details
type TemplateResponse struct {
	Name         string            `json:"name"`
	AbsolutePath string            `json:"absolutePath"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	Descriptions map[string]string `json:"descriptions"`
	RawContent   string            `json:"rawContent"`
}

// HTTPResponse represents the response from executing a template
type HTTPResponse struct {
	StatusCode  int               `json:"statusCode"`
	Status      string            `json:"status"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	ContentType string            `json:"contentType"`
	Duration    int64             `json:"duration"` // milliseconds
	Error       string            `json:"error,omitempty"`
}

// PreviewResponse holds the rendered template fields without making an HTTP request
type PreviewResponse struct {
	URL     string            `json:"url"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
	Error   string            `json:"error,omitempty"`
}

// ConfigInfo returns information about the current configuration
type ConfigInfo struct {
	ConfigPath             string `json:"configPath"`
	TemplatesDir           string `json:"templatesDir"`
	TemplatesDirectoryPath string `json:"templatesDirectoryPath"`
	EnvironmentFile        string `json:"environmentFile"`
}

// GetConfig returns the current configuration
func (a *App) GetConfig() ConfigInfo {
	return ConfigInfo{
		ConfigPath:             a.configPath,
		TemplatesDir:           a.templatesDir,
		TemplatesDirectoryPath: a.config.TemplatesDirectoryPath,
		EnvironmentFile:        a.config.EnvironmentFile,
	}
}

// SaveConfig writes updated configuration fields to the config file
func (a *App) SaveConfig(environmentFile, templatesDirectoryPath string) error {
	configPath := a.configPath
	if configPath == "" {
		configPath = paths.Expand("~/.tp/config.yml")
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}
	cfg := config.Config{
		EnvironmentFile:        environmentFile,
		TemplatesDirectoryPath: templatesDirectoryPath,
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}
	a.configPath = configPath
	a.config = cfg
	a.templatesDir = paths.Expand(templatesDirectoryPath)
	return nil
}

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

		// Load template metadata
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

	err = config.LoadTemplateFiles(a.logger, a.templatesDir, walkFunc)
	if err != nil {
		return nil, err
	}

	// Sort items
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

// ExecuteTemplate executes a template and returns the response
func (a *App) ExecuteTemplate(absolutePath string, overrides map[string]string) HTTPResponse {
	// Convert overrides map to config.Overrides
	var configOverrides config.Overrides
	for k, v := range overrides {
		configOverrides = append(configOverrides, config.Override{Key: k, Value: v})
	}

	resp, err := handlers.ExecuteTemplate(a.logger, absolutePath, a.vars, configOverrides)
	if err != nil {
		return HTTPResponse{
			Error: err.Error(),
		}
	}

	// Convert headers to simple map
	headers := make(map[string]string)
	for k, v := range resp.Headers {
		if len(v) > 0 {
			headers[k] = strings.Join(v, ", ")
		}
	}

	// Format body if JSON
	body := string(resp.Body)
	if strings.Contains(resp.ContentType, "application/json") {
		body = string(pretty.Pretty(resp.Body))
	}

	return HTTPResponse{
		StatusCode:  resp.StatusCode,
		Status:      resp.Status,
		Headers:     headers,
		Body:        body,
		ContentType: resp.ContentType,
		Duration:    resp.Duration.Milliseconds(),
	}
}

// GetVariables returns the current environment variables
func (a *App) GetVariables() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range a.vars {
		if key, ok := k.(string); ok {
			result[key] = v
		}
	}
	return result
}

// ReloadConfig reloads the configuration and environment
func (a *App) ReloadConfig() error {
	cfg, configPath, err := config.LoadOrDefaultConfig(a.logger)
	if err != nil {
		return err
	}
	a.config = cfg
	a.configPath = configPath

	a.templatesDir = paths.Expand(cfg.TemplatesDirectoryPath)
	if a.templatesDir == "" {
		a.templatesDir = paths.Expand("~/.tp/templates")
	}

	_, a.vars = config.LoadEnvironment(a.logger, cfg.EnvironmentFile)
	return nil
}

// CreateTemplate creates a new template file
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

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return "", os.ErrExist
	}

	// Create the template with default content
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

	// Check if folder already exists
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
	// Safety check: ensure path is within templates directory
	if !strings.HasPrefix(absolutePath, a.templatesDir) {
		return os.ErrPermission
	}
	return os.Remove(absolutePath)
}

// RenameItem renames a template file or folder
func (a *App) RenameItem(absolutePath string, newName string) (string, error) {
	// Safety check: ensure path is within templates directory
	if !strings.HasPrefix(absolutePath, a.templatesDir) {
		return "", os.ErrPermission
	}

	// Get the parent directory and construct new path
	parentDir := filepath.Dir(absolutePath)
	newPath := filepath.Join(parentDir, newName)

	// Check if target already exists
	if _, err := os.Stat(newPath); err == nil {
		return "", os.ErrExist
	}

	// Perform the rename
	if err := os.Rename(absolutePath, newPath); err != nil {
		return "", err
	}

	return newPath, nil
}

// MoveItem moves a template or folder to a new parent directory
func (a *App) MoveItem(sourcePath string, targetDir string) (string, error) {
	// Safety check: ensure both paths are within templates directory
	if !strings.HasPrefix(sourcePath, a.templatesDir) {
		return "", os.ErrPermission
	}
	if !strings.HasPrefix(targetDir, a.templatesDir) && targetDir != a.templatesDir {
		return "", os.ErrPermission
	}

	// Get the filename from source
	filename := filepath.Base(sourcePath)
	newPath := filepath.Join(targetDir, filename)

	// Don't move to same location
	if sourcePath == newPath {
		return sourcePath, nil
	}

	// Don't move a folder into itself
	if strings.HasPrefix(targetDir, sourcePath) {
		return "", fmt.Errorf("cannot move folder into itself")
	}

	// Check if target already exists
	if _, err := os.Stat(newPath); err == nil {
		return "", os.ErrExist
	}

	// Perform the move
	if err := os.Rename(sourcePath, newPath); err != nil {
		return "", err
	}

	return newPath, nil
}

// SaveTemplate saves template content
func (a *App) SaveTemplate(absolutePath string, content string) error {
	// Safety check: ensure path is within templates directory
	if !strings.HasPrefix(absolutePath, a.templatesDir) {
		return os.ErrPermission
	}
	return os.WriteFile(absolutePath, []byte(content), 0644)
}

// OpenInEditor opens a file in the system default editor
func (a *App) OpenInEditor(absolutePath string) error {
	// Use wails runtime to open in default application
	// For now, we'll use the system open command
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{"-t", absolutePath} // -t opens in default text editor
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", "", absolutePath}
	default: // linux and others
		cmd = "xdg-open"
		args = []string{absolutePath}
	}

	return exec.Command(cmd, args...).Start()
}

// GetHTTPMethods returns available HTTP methods
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

// GetAppInfo returns application information
func (a *App) GetAppInfo() map[string]string {
	return map[string]string{
		"name":    "tp-gui",
		"version": "0.1.0",
		"built":   time.Now().Format(time.RFC3339),
	}
}

// RefreshVariables reloads environment variables, re-executing any $(command) expansions
// This is useful for refreshing tokens that are fetched via shell commands
func (a *App) RefreshVariables() (map[string]interface{}, error) {
	// Use the environment file from config, or fall back to default paths
	envFile := a.config.EnvironmentFile
	envPath, vars := config.LoadEnvironment(a.logger, envFile)
	a.vars = vars
	a.logger.Infof("RefreshVariables: loaded %d variables from %s (config envFile: %s)", len(vars), envPath, envFile)

	result := a.GetVariables()
	a.logger.Infof("RefreshVariables: returning %d variables", len(result))
	return result, nil
}

// GetEnvironmentFilePath returns the path to the environment/variables file
func (a *App) GetEnvironmentFilePath() string {
	envFile := a.config.EnvironmentFile
	if envFile == "" {
		// Use the same default as the CLI/TUI: ~/.tp/default.env.yml
		envFile = config.DefaultEnvPath
	}
	return paths.Expand(envFile)
}

// OpenEnvironmentFile opens the environment/variables file in the default editor
func (a *App) OpenEnvironmentFile() error {
	return a.OpenInEditor(a.GetEnvironmentFilePath())
}

// GetOverridesFilePath returns the path to the overrides file
func (a *App) GetOverridesFilePath() string {
	return paths.Expand("~/.tp/overrides.yaml")
}

// GetOverrides loads and returns the current overrides from file
func (a *App) GetOverrides() (map[string]string, error) {
	overridesPath := a.GetOverridesFilePath()
	result := make(map[string]string)

	content, err := os.ReadFile(overridesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil // No overrides file yet
		}
		return nil, err
	}

	var rawOverrides map[string]interface{}
	if err := yaml.Unmarshal(content, &rawOverrides); err != nil {
		return nil, err
	}

	for k, v := range rawOverrides {
		if strVal, ok := v.(string); ok {
			result[k] = strVal
		} else {
			// Convert non-string values to string representation
			result[k] = fmt.Sprintf("%v", v)
		}
	}

	return result, nil
}

// SaveOverrides saves overrides to the overrides file
func (a *App) SaveOverrides(overrides map[string]string) error {
	overridesPath := a.GetOverridesFilePath()

	// Ensure directory exists
	dir := filepath.Dir(overridesPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content, err := yaml.Marshal(overrides)
	if err != nil {
		return err
	}

	return os.WriteFile(overridesPath, content, 0644)
}

// OpenOverridesFile opens the overrides file in the default editor
func (a *App) OpenOverridesFile() error {
	overridesPath := a.GetOverridesFilePath()

	// Create the file if it doesn't exist
	if _, err := os.Stat(overridesPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(overridesPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(overridesPath, []byte("# Variable overrides\n# Example: token: my-token-value\n"), 0644); err != nil {
			return err
		}
	}

	return a.OpenInEditor(overridesPath)
}

// PreviewTemplate renders a template with overrides applied but without making an HTTP request
func (a *App) PreviewTemplate(absolutePath string, runtimeOverrides map[string]string) PreviewResponse {
	fileOverrides, err := a.GetOverrides()
	if err != nil {
		a.logger.Warnf("Error loading overrides: %v", err)
		fileOverrides = make(map[string]string)
	}

	mergedOverrides := make(map[string]string)
	for k, v := range fileOverrides {
		mergedOverrides[k] = v
	}
	for k, v := range runtimeOverrides {
		mergedOverrides[k] = v
	}

	var configOverrides config.Overrides
	for k, v := range mergedOverrides {
		configOverrides = append(configOverrides, config.Override{Key: k, Value: v})
	}

	tmpl, err := handlers.RenderTemplate(a.logger, absolutePath, a.vars, configOverrides)
	if err != nil {
		return PreviewResponse{Error: err.Error()}
	}

	return PreviewResponse{
		URL:     strings.TrimSpace(tmpl.Url),
		Body:    strings.TrimSpace(tmpl.Body),
		Headers: tmpl.Headers,
	}
}

// PreviewBody renders a raw body string with overrides applied, without reading from a file
func (a *App) PreviewBody(rawBody string, runtimeOverrides map[string]string) PreviewResponse {
	fileOverrides, err := a.GetOverrides()
	if err != nil {
		a.logger.Warnf("Error loading overrides: %v", err)
		fileOverrides = make(map[string]string)
	}

	mergedOverrides := make(map[string]string)
	for k, v := range fileOverrides {
		mergedOverrides[k] = v
	}
	for k, v := range runtimeOverrides {
		mergedOverrides[k] = v
	}

	var configOverrides config.Overrides
	for k, v := range mergedOverrides {
		configOverrides = append(configOverrides, config.Override{Key: k, Value: v})
	}

	rendered, err := handlers.RenderBodyString(a.logger, rawBody, a.vars, configOverrides)
	if err != nil {
		return PreviewResponse{Body: rendered, Error: err.Error()}
	}
	return PreviewResponse{Body: strings.TrimSpace(rendered)}
}

// ExecuteTemplateWithOverrides executes a template with both env vars and file-based overrides
func (a *App) ExecuteTemplateWithOverrides(absolutePath string, runtimeOverrides map[string]string) HTTPResponse {
	// Load file-based overrides
	fileOverrides, err := a.GetOverrides()
	if err != nil {
		a.logger.Warnf("Error loading overrides: %v", err)
		fileOverrides = make(map[string]string)
	}

	// Merge: file overrides first, then runtime overrides (runtime takes precedence)
	mergedOverrides := make(map[string]string)
	for k, v := range fileOverrides {
		mergedOverrides[k] = v
	}
	for k, v := range runtimeOverrides {
		mergedOverrides[k] = v
	}

	// Convert to config.Overrides
	var configOverrides config.Overrides
	for k, v := range mergedOverrides {
		configOverrides = append(configOverrides, config.Override{Key: k, Value: v})
	}

	resp, err := handlers.ExecuteTemplate(a.logger, absolutePath, a.vars, configOverrides)
	if err != nil {
		return HTTPResponse{
			Error: err.Error(),
		}
	}

	// Convert headers to simple map
	headers := make(map[string]string)
	for k, v := range resp.Headers {
		if len(v) > 0 {
			headers[k] = strings.Join(v, ", ")
		}
	}

	// Format body if JSON
	body := string(resp.Body)
	if strings.Contains(resp.ContentType, "application/json") {
		body = string(pretty.Pretty(resp.Body))
	}

	return HTTPResponse{
		StatusCode:  resp.StatusCode,
		Status:      resp.Status,
		Headers:     headers,
		Body:        body,
		ContentType: resp.ContentType,
		Duration:    resp.Duration.Milliseconds(),
	}
}
