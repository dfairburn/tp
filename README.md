# tp

<img src="https://github.com/dfairburn/tp/assets/47511336/48b17b4e-68ad-4d5b-af26-f60a70dcd0b5" width=100 />


Tp is a command-line utility and library to create and reuse templated http requests.

### [Templates]

Templates are made using Go's template package https://pkg.go.dev/text/template. As you can see from the example below, 
there are some [variables](#variables) that are captured within curly braces. Variables are defined in yaml files that are given as 
an input flag, or defined by config. You may also provide command-line overrides of specific variables.

#### Example
```
===Url
{{ if .UrlAddress }}
    {{ .UrlAddress}}
{{ else }}
    https://a_url.com
{{- end}}

===Method
GET

===Headers
Authorization: Bearer {{ .Token }}
Accept: application/json

===Body
{
    "data": {
        "name": "{{ .Name }}"
    }
}
```

- ===Url
  - The url to make requests to.
- ===Method
    - The http verb to use.
- ===Headers
    - A list of http headers, each one on a newline.
- ===Body
    - The http body to be sent with the request.


### [Variables]

Variables can be input in two forms, either configuring a variables file<sup>1</sup>, or providing them via command-line
overrides.


<sup>1</sup> Should also consider providing directories for more complex variable layouts. Not sure if this comes free 
with yaml parsing or not.

#### Example

### [Config]
