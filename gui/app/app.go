package app

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/paths"
	logging "github.com/sirupsen/logrus"
	"github.com/tidwall/pretty"
)

// App holds the application state
type App struct {
	ctx          context.Context
	logger       *logging.Logger
	config       config.Config
	configPath   string
	vars         map[interface{}]interface{}
	templatesDir string
}

// TemplateItem represents a template or directory for the frontend
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

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	a.logger = logging.New()
	a.logger.SetLevel(logging.InfoLevel)
	if logFile, err := os.OpenFile(paths.Expand("~/.tp/tp-gui.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		a.logger.Out = logFile
	} else {
		a.logger.Out = io.Discard
	}

	cfg, configPath, err := config.LoadOrDefaultConfig(a.logger)
	if err != nil {
		a.logger.Warnf("Error loading config: %v", err)
	}
	a.config = cfg
	a.configPath = configPath

	a.templatesDir = paths.Expand(cfg.TemplatesDirectoryPath)
	if a.templatesDir == "" {
		a.templatesDir = paths.Expand("~/.tp/templates")
	}

	_, a.vars = config.LoadEnvironment(a.logger, cfg.EnvironmentFile)
}

// ── Shared helpers ────────────────────────────────────────────────────────────

// mapToConfigOverrides converts a plain map to the config.Overrides slice type.
func mapToConfigOverrides(m map[string]string) config.Overrides {
	var out config.Overrides
	for k, v := range m {
		out = append(out, config.Override{Key: k, Value: v})
	}
	return out
}

// mergeOverrides loads file-based overrides, merges with the given runtime overrides
// (runtime takes precedence), and converts to config.Overrides.
func (a *App) mergeOverrides(runtimeOverrides map[string]string) config.Overrides {
	fileOverrides, err := a.GetOverrides()
	if err != nil {
		a.logger.Warnf("Error loading overrides: %v", err)
		fileOverrides = make(map[string]string)
	}
	merged := make(map[string]string, len(fileOverrides)+len(runtimeOverrides))
	for k, v := range fileOverrides {
		merged[k] = v
	}
	for k, v := range runtimeOverrides {
		merged[k] = v
	}
	return mapToConfigOverrides(merged)
}

// buildHTTPResponse converts raw HTTP response data into the HTTPResponse DTO.
func buildHTTPResponse(statusCode int, status string, header http.Header, body []byte, duration int64) HTTPResponse {
	headers := make(map[string]string, len(header))
	for k, v := range header {
		if len(v) > 0 {
			headers[k] = strings.Join(v, ", ")
		}
	}
	contentType := header.Get("Content-Type")
	bodyStr := string(body)
	if strings.Contains(contentType, "application/json") {
		bodyStr = string(pretty.Pretty(body))
	}
	return HTTPResponse{
		StatusCode:  statusCode,
		Status:      status,
		Headers:     headers,
		Body:        bodyStr,
		ContentType: contentType,
		Duration:    duration,
	}
}

