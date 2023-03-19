package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/vars"
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
func Use(path string, config config.Config) error {
	fmt.Println(config.VariableDefinitionFile)
	v := vars.LoadVars(config.VariableDefinitionFile)

	tp := template.Must(template.ParseFiles(path))

	var buf bytes.Buffer
	err := tp.Execute(&buf, v)

	req := NewRequest(buf)
	r, err := req.toHttp()
	if err != nil {
		return err
	}

	cli := http.Client{}
	resp, err := cli.Do(r)
	if err != nil {
		return err
	}

	fmt.Println(resp)
	return err
}

type Request struct {
	Method  string
	Headers map[string]string
	Body    string
	Url     string
}

func NewRequest(b bytes.Buffer) Request {
	r := Request{
		Headers: make(map[string]string),
	}
	for _, key := range keys {
		sub := strings.Split(b.String(), key)
		if len(sub) < 1 {
			continue
		}
		k := toKey(key)
		v := findNextSection(sub[1])[0]

		buildRequest(&r, k, v)
	}

	return r
}

func (r Request) toHttp() (*http.Request, error) {
	reqBody := bytes.NewBufferString(r.Body)
	return http.NewRequestWithContext(context.Background(), r.Method, r.Url, reqBody)
}

func buildRequest(r *Request, k string, v string) *Request {
	switch k {
	case url:
		trim(&v)
		r.Url = v
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
		r.Body = v
	}

	return r
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
