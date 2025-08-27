package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	httpurl "net/url"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/dfairburn/tp/config"
	"gopkg.in/yaml.v3"

	"github.com/tidwall/pretty"

	logging "github.com/sirupsen/logrus"
)

type Template struct {
	Descriptions map[string]string `yaml:"descriptions"`
	Url          string            `yaml:"url"`
	Method       string            `yaml:"method"`
	Headers      map[string]string `yaml:"headers"`
	Body         string            `yaml:"body"`
}

// Use gets a filepath and "uses" that template
func Use(logger *logging.Logger, templateFile string, vars map[interface{}]interface{}, overrides config.Overrides, rawOutput bool) error {
	_, err := os.Stat(templateFile)
	if err != nil {
		return err
	}

	templateName := path.Base(templateFile)

	tmpl, err := template.New(templateName).
		Funcs(templateFuncs(logger)).
		ParseFiles(templateFile)
	if err != nil {
		return err
	}
	overridden := Override(vars, overrides)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, templateName, overridden)
	if err != nil {
		return err
	}

	if buf.Len() < 1 {
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

	ct := resp.Header.Get(http.CanonicalHeaderKey("Content-Type"))
	if strings.Contains(ct, "application/json") && !rawOutput {
		respBody = formatResponse(respBody)
	}

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

	q := u.Query()
	u.RawQuery = q.Encode()

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

func formatResponse(respBody []byte) []byte {
	// pretty-print json output
	body := pretty.Pretty(respBody)
	o, _ := os.Stdout.Stat()
	if (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		// colorize when Stdout is a terminal
		body = pretty.Color(body, nil)
	}
	return body
}

func templateFuncs(logger *logging.Logger) template.FuncMap {
	return template.FuncMap{
		"default": func(defaultValue string, value any) string {
			if value == nil {
				return defaultValue
			}
			return fmt.Sprintf("%s", value)
		},
		"optional": func(format string, value any) string {
			str, ok := value.(string)
			if !ok || str == "" {
				return ""
			}
			return fmt.Sprintf(format, value)
		},
		"timestamp": func(value any) string {
			str, ok := value.(string)
			if !ok || str == "" {
				// user chose to leave this empty
				return ""
			}
			if _, err := time.Parse(time.RFC3339, str); err == nil {
				// user provided a timestamp, just use it
				return str
			}
			if str == "now" {
				// now is a special keyword for the current time
				return time.Now().UTC().Format(time.RFC3339)
			}
			if dur, err := time.ParseDuration(str); err == nil {
				// a duration string relative to now
				return time.Now().Add(dur).Format(time.RFC3339)
			}

			logger.Warnf("template function 'timestamp' received unexpected value: %v", value)
			return ""
		},
	}
}
