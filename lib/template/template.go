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
	return &Templater{
		Cutter:   cutter,
		Sections: sectionKeys,
	}
}

type Section struct {
	Key   Key
	Value string
}

type Key string

func NewKey(s string) Key {
	nonAlphaRegex := regexp.MustCompile(`[^a-zA-Z]+`)
	sanitised := strings.ToLower(nonAlphaRegex.ReplaceAllString(s, ""))

	return Key(sanitised)
}

// This is now repurposed as the 'cutter'
func findNextSection(s string) []string {
	return strings.Split(s, "===")
}

func NewSection() Section {
	return Section{}
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
