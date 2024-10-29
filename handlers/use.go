package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	httpurl "net/url"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/dfairburn/tp/config"

	logging "github.com/sirupsen/logrus"
)

type Template struct {
	Url     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
}

// Use gets a filepath and "uses" that template
func Use(logger *logging.Logger, templateFile string, vars map[interface{}]interface{}, overrides config.Overrides) error {
	_, err := os.Stat(templateFile)
	if err != nil {
		return err
	}

	tp := template.Must(template.ParseFiles(templateFile)) // .Option("missingkey=error")
	overridden := Override(vars, overrides)

	var buf bytes.Buffer
	err = tp.Execute(&buf, overridden)
	if err != nil {
		return err
	}

	if len(buf.Bytes()) < 1 {
		return errors.New("unexpected 0 length from executing template")
	}

	tmp := &Template{}
	err = yaml.Unmarshal(buf.Bytes(), tmp)
	if err != nil {
		return err
	}

	req, err := NewRequest(tmp)
	if err != nil {
		return err
	}

	r, err := req.toHttp()
	if err != nil {
		return err
	}

	logger.Println("method:", r.Method)
	logger.Println("url:", r.URL)
	logger.Println("body:", r.Body)
	logger.Println("headers:", r.Header)

	cli := http.Client{}

	resp, err := cli.Do(r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Println(err)
		return err
	}
	logger.Println(resp)
	_, err = io.WriteString(os.Stdout, string(respBody))
	if err != nil {
		logger.Println(err)
		return err
	}

	return nil
}

type Request struct {
	Method  string
	Headers map[string]string
	Body    string
	Url     *httpurl.URL
}

func NewRequest(tmp *Template) (*Request, error) {
	strippedBody := strings.TrimSpace(tmp.Body)
	strippedURL := strings.TrimSpace(tmp.Url)
	u, err := httpurl.Parse(strippedURL)
	if err != nil {
		return nil, err
	}

	r := &Request{
		Url:     u,
		Method:  tmp.Method,
		Headers: tmp.Headers,
		Body:    strippedBody,
	}

	return r, nil
}

func (r Request) toHttp() (*http.Request, error) {
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
		reqBody := bytes.NewBufferString(r.Body)
		jsonReq, err := http.NewRequestWithContext(context.Background(), r.Method, r.Url.String(), reqBody)
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

func Override(vars map[interface{}]interface{}, overrides config.Overrides) any {
	if len(vars) == 0 {
		return overrides.ToMap()
	}

	for _, override := range overrides {
		vars[override.Key] = override.Value
	}

	return vars
}
