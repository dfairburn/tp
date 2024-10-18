# tp

<img src="https://github.com/dfairburn/tp/assets/47511336/48b17b4e-68ad-4d5b-af26-f60a70dcd0b5" width=200 />

Tp is a command-line utility to create and reuse templated http requests.

---
## Quickstart

### Installation

TODO

### Creating a template

To create your first template, you can use the `tp open` command:

```shell
tp open example
```

This should open up your default configured editor using the `$EDITOR` environment variable (which defaults to `vim`). You can then populate the default template with the following values:

```
===Url
https://jsonplaceholder.typicode.com/users

===Method
GET

===Headers

===Body
```

Once written, a file will have been written to `~/.tp/templates/example.tmpl` and you'll be ready to use it.

#### Notes

* To find out more about the `tp open` subcommand, please see [the tp use command](#tp-use)
* To find out more about templates, please see [the templates section](#templates)

### Executing a template

Now that a template has been set up, you can then execute the template with the following command:

```shell
tp use example
```

```json5
{
  // some json response...
}
```


#### Notes
* There are more ways to execute the `tp use` subcommand, to read about the other options, head to [the tp use command](tp-use)


### Using Variables

This is all well and good, but the whole point of this is to template out these requests to make them easier to use and configure. So let's add some variables.

Execute the following command to open up the varaibles file in your default editor:

```shell
tp vars
```

Which should open up an empty yaml file with the filepath `~/.tp/vars` and should look like this:

```yaml
---
```


// TODO - finish the readme :D

## Subcommands

### tp use
### tp open
### tp vars
### tp list
### tp completion
### tp config

### Templates

tp has the notion of "templates", which are structured files that hold data to construct HTTP requests. The template structure is as follows:

```
===Url
// the target url of the HTTP request

===Method
// the HTTP method to be used

===Headers
// any additional headers to be sent with the request

===Body
// the data body to be sent with the HTTP request
```

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
