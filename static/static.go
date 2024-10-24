package static

import (
	_ "embed"
)

//go:embed template.yml
var BaseTemplate []byte

const YamlRegex = "(?:.*\\.yaml$|.*\\.yml$)"
