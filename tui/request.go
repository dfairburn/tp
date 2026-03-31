package tui

import (
	"net/http"
	"strings"
	"time"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/handlers"
	logging "github.com/sirupsen/logrus"
	"github.com/tidwall/pretty"
)

// RequestResult holds the result of an HTTP request for the TUI
type RequestResult struct {
	StatusCode  int
	Status      string
	Headers     http.Header
	Body        string
	Duration    time.Duration
	Error       error
	ContentType string
	Size        int
	Request     *RequestInfo
}

// RequestInfo holds info about the sent request
type RequestInfo struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// ExecuteRequest executes an HTTP request from a template using the handlers package
func ExecuteRequest(logger *logging.Logger, templateFile string, vars map[interface{}]interface{}, overrides config.Overrides) *RequestResult {
	resp, err := handlers.ExecuteTemplate(logger, templateFile, vars, overrides)
	if err != nil {
		return &RequestResult{
			Error: err,
		}
	}

	// Format JSON if applicable
	body := resp.Body
	if strings.Contains(resp.ContentType, "application/json") {
		body = pretty.Pretty(body)
	}

	result := &RequestResult{
		StatusCode:  resp.StatusCode,
		Status:      resp.Status,
		Headers:     resp.Headers,
		Body:        string(body),
		Duration:    resp.Duration,
		ContentType: resp.ContentType,
		Size:        len(resp.Body),
	}

	// Convert request info if available
	if resp.Request != nil {
		result.Request = &RequestInfo{
			Method:  resp.Request.Method,
			URL:     resp.Request.Url.String(),
			Headers: resp.Request.Headers,
			Body:    resp.Request.Body,
		}
	}

	return result
}
