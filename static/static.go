package static

import (
	_ "embed"
)

//go:embed template.yml
var DefaultTemplate []byte

//go:embed config.yml
var DefaultConfig []byte

const YamlRegex = "(?:.*\\.yaml$|.*\\.yml$)"
