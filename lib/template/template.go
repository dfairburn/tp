package template

import (
	"regexp"
	"strings"
)

type Templater struct {
	Sections []Section
	Cutter   func(s string) []string
}

func NewTemplate(keys []string, cutter func(s string) []string) *Templater {
	var sections []Section

	for _, k := range keys {
		sect := NewSection(k)
		sections = append(sections, sect)
	}

	return &Templater{
		Cutter:   cutter,
		Sections: sections,
	}
}

type Section struct {
	Key   Key
	Value string
}

type Key string

// This is now repurposed as the 'cutter'
func findNextSection(s string) []string {
	return strings.Split(s, "===")
}

func NewSection(s string) Section {
	nonAlphaRegex := regexp.MustCompile(`[^a-zA-Z]+`)
	sanitised := strings.ToLower(nonAlphaRegex.ReplaceAllString(s, ""))
	key := Key(sanitised)

	return Section{
		Key: key,
	}
}

func SectionsFromString() []Section {
	return nil
}

//for _, key := range keys {
//sub := strings.Split(b.String(), key)
//if len(sub) < 1 {
//continue
//}
//k := toKey(key)
//v := findNextSection(sub[1])[0]
