package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	httpurl "net/url"
	"regexp"
	"strings"
	"text/template"

	"github.com/dfairburn/tp/config"
	logging "github.com/sirupsen/logrus"
)

const (
	urlKey     = "===Url"
	url        = "url"
	methodKey  = "===Method"
	method     = "method"
	headersKey = "===Headers"
	headers    = "headers"
	bodyKey    = "===Body"
	body       = "body"
)

var (
	keys = []string{methodKey, urlKey, headersKey, bodyKey}
)

// Use gets a filepath and "uses" that template
func Use(logger *logging.Logger, templateFile string, vars map[interface{}]interface{}, overrides config.Overrides) error {
	fmt.Println(vars)
	fmt.Println(templateFile)
	fmt.Println(overrides)
	tp := template.Must(template.ParseFiles(templateFile))
	overridden := Override(vars, overrides)

	var buf bytes.Buffer
	err := tp.Execute(&buf, overridden)
	if err != nil {
		return err
	}

	if len(buf.Bytes()) < 1 {
		return errors.New("template not executed, missing variables")
	}

	req, err := NewRequest(buf)
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

	logger.Println(resp)
	logger.Println(err)
	return err
}

type Request struct {
	Method  string
	Headers map[string]string
	Body    string
	Url     *httpurl.URL
}

func NewRequest(b bytes.Buffer) (*Request, error) {
	r := &Request{
		Headers: make(map[string]string),
	}

	// get data for each section
	for _, key := range keys {
		sub := strings.Split(b.String(), key)
		if len(sub) < 2 {
			continue
		}
		k := toKey(key)
		v := findNextSection(sub[1])[0]

		err := buildRequest(r, k, v)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r Request) toHttp() (*http.Request, error) {
	reqBody := bytes.NewBufferString(r.Body)
	return http.NewRequestWithContext(context.Background(), r.Method, r.Url.String(), reqBody)
}

func buildRequest(r *Request, k string, v string) error {
	switch k {
	case url:
		trim(&v)
		parsed, err := httpurl.Parse(v)
		if err != nil {
			return fmt.Errorf("cannot parse %s as a url: %w", v, err)
		}
		r.Url = parsed
	case method:
		trim(&v)
		r.Method = v
	case headers:
		h := strings.Split(v, "\n")
		for _, header := range h {
			if len(header) == 0 {
				continue
			}

			// TODO: Add some error handling
			parts := strings.Split(header, ":")
			if len(parts) < 2 {
				continue
			}

			key, val := parts[0], parts[1]
			trim(&key, &val)
			r.Headers[key] = val
		}
	case body:
		if strings.TrimSpace(v) != "" {
			r.Body = v
		}
	}

	return nil
}

func trim(s ...*string) {
	for _, str := range s {
		newStr := strings.TrimSpace(*str)
		*str = newStr
	}
}

func findNextSection(s string) []string {
	return strings.Split(s, "===")
}

func toKey(s string) string {
	nonAlphaRegex := regexp.MustCompile(`[^a-zA-Z]+`)
	return strings.ToLower(nonAlphaRegex.ReplaceAllString(s, ""))
}

func Override(vars map[interface{}]interface{}, overrides config.Overrides) any {
	if len(vars) == 0 {
		return overrides.ToMap()
	}

	matchMap := lowerCaseMap(vars)

	for _, override := range overrides {
		// lower-case key to match with lower-cased var map
		lowerKey := strings.ToLower(override.Key)
		if key, ok := matchMap[lowerKey]; ok {
			vars[key] = override.Value
		}
	}

	return vars
}

func lowerCaseMap(y map[interface{}]interface{}) map[string]string {
	keyToLowercaseMappings := make(map[string]string)
	for k, _ := range y {
		if key, ok := k.(string); ok {
			keyToLowercaseMappings[strings.ToLower(key)] = key
		}
	}

	return keyToLowercaseMappings
}
