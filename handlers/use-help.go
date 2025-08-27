package handlers

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

func ParseUsages(content []byte) ([]Usage, error) {
	usageRegex := regexp.MustCompile(`\{\{([^\}]+)\}\}`)
	varUsage := regexp.MustCompile(`^\s*\.(\S+)\s*$`)
	optionalUsage := regexp.MustCompile(`^\s*optional.*\.(\S+)\s*$`)
	timestampUsage := regexp.MustCompile(`^\s*timestamp\s+\.(\S+)\s*$`)
	defaultUsage := regexp.MustCompile(`^\s*default\s+\.(\S+)\s+"(\w+)"\s*$`)

	uses := []Usage{}
	for _, match := range usageRegex.FindAllStringSubmatch(string(content), -1) {
		if len(match) != 2 {
			continue
		}
		useString := match[1]
		switch {
		case varUsage.MatchString(useString):
			match := varUsage.FindStringSubmatch(useString)
			uses = append(uses, VarUsage{ref: match[1]})
		case optionalUsage.MatchString(useString):
			match := optionalUsage.FindStringSubmatch(useString)
			uses = append(uses, OptionalUsage{ref: match[1]})
		case timestampUsage.MatchString(useString):
			match := timestampUsage.FindStringSubmatch(useString)
			uses = append(uses, TimestampUsage{ref: match[1]})
		case defaultUsage.MatchString(useString):
			match := defaultUsage.FindStringSubmatch(useString)
			uses = append(uses, DefaultUsage{ref: match[1], defaultVal: match[2]})
		}
	}
	return uses, nil
}

func GenerateTemplateUsage(content []byte) (string, error) {
	uses, err := ParseUsages(content)
	if err != nil {
		return "", err
	}

	template := Template{}
	err = yaml.Unmarshal(content, &template)
	if err != nil {
		return "", err
	}

	var result string
	for _, use := range uses {
		desc := template.Descriptions[use.Name()]
		result += fmt.Sprintf("  %-15s%-20s%-25s\n", use.Name(), use.Extra(), desc)
	}
	return result, nil
}

type Usage interface {
	Name() string
	Extra() string
}

type VarUsage struct {
	ref string
}

func (v VarUsage) Name() string {
	return v.ref
}

func (v VarUsage) Extra() string {
	return ""
}

type OptionalUsage struct {
	ref string
}

func (v OptionalUsage) Name() string {
	return v.ref
}
func (v OptionalUsage) Extra() string {
	return "(optional)"
}

type DefaultUsage struct {
	ref        string
	defaultVal string
}

func (v DefaultUsage) Name() string {
	return v.ref
}
func (v DefaultUsage) Extra() string {
	return fmt.Sprintf("(default: %s)", v.defaultVal)
}

type TimestampUsage struct {
	ref string
}

func (v TimestampUsage) Name() string {
	return v.ref
}
func (v TimestampUsage) Extra() string {
	return "(timestamp)"
}
