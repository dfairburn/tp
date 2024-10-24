# tp

<img src="https://github.com/dfairburn/tp/assets/47511336/48b17b4e-68ad-4d5b-af26-f60a70dcd0b5" width=200 />

Tp is a command-line utility to create and reuse templated http requests. Tp was created as an alternative to 
GUI-based API dev tools. The driving motivation for tp was to build something that is:
* simple to use
* able to open all your templates and environment variables in any editor of your choice
* modular and reusable
* configurable at the point of execution (providing overrides on the command-line for variables that could change frequently)

---
## Quickstart

### Installation

#### Go

Requirements:
  * Go v1.20

### Creating a template

To create your first template, you can use the `tp open` command:

```shell
tp open example
```

This should open up your default configured editor using the `$EDITOR` environment variable (which defaults to `vim`). You can then populate the default template with the following values:

```yaml
url: https://jsonplaceholder.typicode.com/users
method: GET
headers:
body:
```

Once written, a file will have been written to `~/.tp/templates/example.yml` and you'll be ready to use it.

#### Notes

* To find out more about the `tp open` subcommand, please see [the tp use command](#tp-use)
* To find out more about templates, please see [the templates section](#templates)

### Executing a template

Now that a template has been set up, you can then execute the template with the following command:

```shell
tp use example
```

```json5
// example response from jsonplaceholder.typicode.com
{
  "id": 1,
  "name": "Leanne Graham",
  "username": "Bret",
  "email": "Sincere@april.biz",
  "address": {
    // ...
  },
  "phone": "1-770-736-8031 x56442",
  "website": "hildegard.org",
  "company": {
    // ...
  }
}
```

#### Notes
* There are more ways to execute the `tp use` subcommand, to read about the other options, head to [the tp use command](tp-use)


### Using Variables

This is all well and good, but the whole point of this is to template out these requests to make them easier to use, 
configure and mutate on the fly. So let's add some variables.

Execute the following command to open up the varaibles file in your default editor.
On first run, this should open up an empty yaml file with the filepath `~/.tp/vars`.

```shell
tp vars
```


Let's add a new variable to this file like so:

**NOTE:** Go's templating package does not support `-`, hence why it is not used here!
```yaml
---

content_type: application/json
user: 1
```

We can then edit the template file by using `tp open example` and making the template file look as follows:

```yaml
url: https://jsonplaceholder.typicode.com/users/{{ .user }}
method: GET
headers:
  - Content-Type: {{ .content_type }}
body:
```

You can then run `tp use example` to execute this template and with the provided variables.

### Variables currently support
#### Overriding

You can override template variables by passing the `-o` flag, to the `use` command. You can read more about this in the 
[tp use](#tp-use) section.
#### Nesting

You are able to nest variables (like you would in the standard yaml format) and use them in your templates:
```yaml
users:
  user_1: test_username_guid
```
which in the template can be referenced like so:
```yaml
url: https://jsonplaceholder.typicode.com/users/{{ .users.user_1 }}
```

#### Shell expansion
You're able to pass in commands that are executed by the shell to be stored inside of variables to be reused in your templates:

```yaml
token: $(get-api-token)
```

Which can be referenced as you would a normal/nested variable:
```yaml
headers:
  - Authorization: Bearer {{ .token }}
```

This gets evaluated by your shell (defined by the `$SHELL` env var) on the load of the variable file.

#### Notes on Variables

The Variables file is defined in yaml, however the go `templating` package doesn't seem to like varaibles
that have `-` in the name, `A-Za-z`, `0-9` and `_` are valid for variable names.

## Subcommands
#### Global Flags

| Short | Long     | Description                                                                             |
|-------|----------|-----------------------------------------------------------------------------------------|
|       | --config | An override for the location of the config file to use (defaults to: ~/.tp/config.yml)  |
|       | --vars   | An override for the location of the variables file to use (defaults to: ~/.tp/vars.yml) |
|       | --debug  | Redirects all logging to STDOUT                                                         |
|       | --log    | An override for the destination of the logfile to be written to                         |
| -h    | --help   | Displays the help text for the `use` command.                                           |

### tp use
`tp use` takes a given template, interpolates the variables defined in the configured variable file, takes
in any defined overrides for these variables, and sends an http request to the url defined in the template

You can either give `tp use` a filename, or you can execute it without a filename and tp will open a fuzzyfind within
your configured templates directory.

**Note**
The default templates directory is `~/.tp/templates`, however you can define your own within the `~/.tp/config.yml` file.

**Nesting**
`tp use` supports the execution and searching of templates nested in directories, for example, given the directory structure:

```
templates/
  jsonapi/
    example.yml
```

you can execute:

```shell
tp use jsonapi/example
```

and tp will execute the existing template at `~/.tp/templates/jsonapi/example.yml`

**Example:**

`tp use example` will execute the template defined at: `~/.tp/templates/example` (if using the default templates directory location)

**Shell Completion:**

`tp use` supports tab completion, and you can find out how to set this up for your shell in the 
[tp completion](#tp-completion) section

Flags:

| Short | Long        | Description                                                                    |
|-------|-------------|--------------------------------------------------------------------------------|
| -o    | --overrides | A list of variable overrides, either comma-separated or by repeating the flag. |
| -h    | --help      | Displays the help text for the `use` command.                                  |

Example `-o` usage:

```shell
tp use -o user:"1" -o content_type:"application/json" -o time:$(time)
```

```shell
tp use --overrides user:"1" --overrides content_type:"application/json" --overrides time:$(time)
```

```shell
tp use --overrides user:"1",content_type:"application/json",time:$(time)
```

### tp open

`tp open` takes a filename as a command and either opens a new template, with the default template syntax, or opens
an already existing template.

**Note**
The default templates directory is `~/.tp/templates`, however you can define your own within the `~/.tp/config.yml` file.

### tp vars

`tp vars` opens the variable file defined in your tp config (default: `~/.tp/config.yml`) in your configured editor. The
editor is chosen based on what your `$EDITOR` environment variable is set to

**Note**

If your `$EDITOR` env var is not set, the default editor will be `vim`

### tp list

`tp list` prints the absolute path for all of your templates in your configured templates dir to STDOUT

### tp completion

`tp completion` prints the instructions to follow for installing shell completions. For completeness, I'll also state
these below:

To load completions:

**Bash:**
```shell
source <(tp completion bash)
```
To load completions for each session, execute once:

**Linux:**
```shell
tp completion bash > /etc/bash_completion.d/tp
```
**macOS:**
```shell
tp completion bash > $(brew --prefix)/etc/bash_completion.d/tp
```

---

**Zsh:**

If shell completion is not already enabled in your environment,
you will need to enable it.  You can execute the following once:

```shell
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

To load completions for each session, execute once:
```shell
tp completion zsh > "${fpath[1]}/_tp"
```

You will need to start a new shell for this setup to take effect.

---
  
**fish:**

```shell
tp completion fish | source
```
To load completions for each session, execute once:
```shell
tp completion fish > ~/.config/fish/completions/tp.fish
```

---

**PowerShell:**

```
PS> tp completion powershell | Out-String | Invoke-Expression
```

To load completions for every new session, run:
```
PS> tp completion powershell > tp.ps1
```
and source this file from your PowerShell profile.

### tp config

`tp config` loads the config file defined at `~/.tp/config.yml`. This is where the default locations for
your templates directory and variables file are stored. The default `~/.tp/config.yml` looks as follows:

```yaml
---

variableDefinitionFile: "~/.tp/vars.yml"
templatesDirectoryPath: "~/.tp/templates"
```

### Templates

tp has the notion of "templates", which are yaml files that hold data to construct HTTP requests. The template structure is as follows:

```yaml
# the target url of the HTTP request
url: 
# the HTTP method to be used
method: 
# any additional headers to be sent with the request
headers:
  -
# the data body to be sent with the HTTP request
body:
```

Templates are made using [Go's template package](https://pkg.go.dev/text/template). As you can see from the example below, 
there are some [variables](#variables) that are captured within curly braces. Variables are defined in yaml files that are given as 
an input flag, or defined by config. You may also provide command-line overrides of specific variables.

#### Example
```yaml
url: >
{{ if .UrlAddress }}
{{ .UrlAddress}}
{{ else }}
    https://jsonplaceholder.typicode.com/users
{{- end}}


method: GET

headers:
  - Authorization: Bearer {{ .Token }}
  - Accept: application/json

body: >
{
  "data": {
    "name": "{{ .Name }}"
  }
}
```

### Variables

Variables can be input in two forms, either configuring a variables file<sup>1</sup>, or providing them via command-line
overrides.


<sup>1</sup> Should also consider providing directories for more complex variable layouts. Not sure if this comes free 
with yaml parsing or not.

### Config
