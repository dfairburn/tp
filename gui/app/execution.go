package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	httpurl "net/url"
	"strings"
	"time"

	"github.com/dfairburn/tp/handlers"
)

// httpClient is a shared client for all GUI-initiated requests, enabling TCP connection reuse.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// ExecuteTemplate executes a template and returns the response
func (a *App) ExecuteTemplate(absolutePath string, overrides map[string]string) HTTPResponse {
	resp, err := handlers.ExecuteTemplate(a.logger, absolutePath, a.vars, mapToConfigOverrides(overrides))
	if err != nil {
		return HTTPResponse{Error: err.Error()}
	}
	return buildHTTPResponse(resp.StatusCode, resp.Status, resp.Headers, resp.Body, resp.Duration.Milliseconds())
}

// ExecuteTemplateWithBodyAndOverrides executes a template using a caller-supplied body
// instead of the body defined in the template file. URL, method, and headers are still
// rendered from the template with overrides applied.
func (a *App) ExecuteTemplateWithBodyAndOverrides(absolutePath string, runtimeOverrides map[string]string, bodyOverride string) HTTPResponse {
	startTime := time.Now()

	tmp, err := handlers.RenderTemplate(a.logger, absolutePath, a.vars, a.mergeOverrides(runtimeOverrides))
	if err != nil {
		return HTTPResponse{Error: err.Error()}
	}

	tmp.Body = bodyOverride

	req, err := handlers.NewRequest(tmp)
	if err != nil {
		return HTTPResponse{Error: err.Error()}
	}

	r, err := toHTTPRequest(req)
	if err != nil {
		return HTTPResponse{Error: err.Error()}
	}

	resp, err := httpClient.Do(r)
	if err != nil {
		return HTTPResponse{Error: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return HTTPResponse{Error: err.Error()}
	}

	return buildHTTPResponse(resp.StatusCode, resp.Status, resp.Header, body, time.Since(startTime).Milliseconds())
}

// ExecuteTemplateWithOverrides executes a template with both env vars and file-based overrides
func (a *App) ExecuteTemplateWithOverrides(absolutePath string, runtimeOverrides map[string]string) HTTPResponse {
	resp, err := handlers.ExecuteTemplate(a.logger, absolutePath, a.vars, a.mergeOverrides(runtimeOverrides))
	if err != nil {
		return HTTPResponse{Error: err.Error()}
	}
	return buildHTTPResponse(resp.StatusCode, resp.Status, resp.Headers, resp.Body, resp.Duration.Milliseconds())
}

// PreviewTemplate renders a template with overrides applied but without making an HTTP request
func (a *App) PreviewTemplate(absolutePath string, runtimeOverrides map[string]string) PreviewResponse {
	tmpl, err := handlers.RenderTemplate(a.logger, absolutePath, a.vars, a.mergeOverrides(runtimeOverrides))
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
	rendered, err := handlers.RenderBodyString(a.logger, rawBody, a.vars, a.mergeOverrides(runtimeOverrides))
	if err != nil {
		return PreviewResponse{Body: rendered, Error: err.Error()}
	}
	return PreviewResponse{Body: strings.TrimSpace(rendered)}
}

// toHTTPRequest converts a handlers.Request into a standard *http.Request.
// Mirrors the unexported handlers.Request.toHttp() method so the GUI can build
// requests without modifying the handlers package.
func toHTTPRequest(r *handlers.Request) (*http.Request, error) {
	var req *http.Request

	switch r.Headers["Content-Type"] {
	case "application/x-www-form-urlencoded":
		data := httpurl.Values{}
		for _, d := range strings.Split(r.Body, " ") {
			values := strings.Split(d, "=")
			if len(values) != 2 {
				return nil, errors.New("expected key and value in form body")
			}
			data.Set(values[0], values[1])
		}
		formReq, err := http.NewRequestWithContext(context.Background(), r.Method, r.Url.String(), strings.NewReader(data.Encode()))
		if err != nil {
			return nil, err
		}
		req = formReq
	default:
		jsonReq, err := http.NewRequestWithContext(context.Background(), r.Method, r.Url.String(), bytes.NewBufferString(r.Body))
		if err != nil {
			return nil, err
		}
		req = jsonReq
	}

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	return req, nil
}
