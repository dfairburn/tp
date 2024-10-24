package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	logging "github.com/sirupsen/logrus"
)

func LoadVars(logger *logging.Logger, paths ...string) (string, map[interface{}]interface{}) {
	y := make(map[interface{}]interface{})

	// this gives precedent to paths passed in via config and flags, then processes the default file paths
	paths = append(paths, buildVarPaths().paths...)

	f, path, err := tryFiles(logger, paths...)
	if err != nil {
		logger.Errorf("unable to use var files, error: %v", err)
		return path, y
	}

	data, err := io.ReadAll(f)
	if err != nil {
		logger.Fatalf("error: %v", err)
	}

	err = yaml.Unmarshal(data, &y)
	if err != nil {
		logger.Fatalf("error: %v", err)
	}

	variables := expandVars(logger, y)

	return path, variables
}

func expandVars(logger *logging.Logger, y map[any]any) map[any]any {
	expandedMap := make(map[any]any)

	for key, value := range y {
		switch value.(type) {
		case string:
			v := value.(string)
			re := regexp.MustCompile("\\$\\((?P<command>.*)\\)")
			result := make(map[string]string)
			if !re.MatchString(v) {
				expandedMap[key] = value
				continue
			}

			match := re.FindStringSubmatch(v)
			for i, name := range re.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = match[i]
				}
			}

			cmd, ok := result["command"]
			if !ok {
				expandedMap[key] = value
				continue
			}

			shell := os.Getenv("SHELL")
			e := exec.Command(shell, "-c", cmd)
			var out strings.Builder
			e.Stdout = &out
			err := e.Run()
			if err != nil {
				logger.
					WithField("shell", shell).
					WithField("command", cmd).
					WithError(err).
					Fatal("error expanding variable file, executing command")
			}

			expanded := strings.TrimSuffix(out.String(), "\n")
			expandedMap[key] = expanded
		case map[string]interface{}:
			// need to cast the map back into a map[interface{}]interface{} to feed back into
			// the expandVars func to be able to expand nested vars
			m := make(map[interface{}]interface{})
			v := value.(map[string]interface{})
			for k, vv := range v {
				m[k] = vv
			}

			expanded := expandVars(logger, m)
			expandedMap[key] = expanded
			continue
		default:
			expandedMap[key] = value
			continue
		}
	}

	return expandedMap
}

type Override struct {
	Key   string
	Value string
}

type Overrides []Override

func (o *Overrides) ToMap() map[interface{}]interface{} {
	m := make(map[interface{}]interface{})
	if len(*o) == 0 {
		return m
	}

	for _, ov := range *o {
		m[ov.Key] = ov.Value
	}

	return m
}

func newOverride(o string, logger *logging.Logger) (Override, error) {
	ovr := strings.Split(o, ":")
	if len(ovr) < 2 {
		return Override{}, fmt.Errorf("overrides should be given in the form of 'name:value'. "+
			"No colon was found to split name and value: %s", o)
	}
	if len(ovr) > 2 {
		logger.Warnf("extra colon used in override. If these are meant to be separate overrides, please add"+
			" an extra colon to separate the overrides: %s", o)
	}

	joined := strings.Join(ovr[1:], ":")

	return Override{
		Key:   ovr[0],
		Value: joined,
	}, nil
}

func ValidateOverrides(o []string, logger *logging.Logger) ([]Override, error) {
	var overrideSlice []Override
	var errs []error
	for _, override := range o {
		validOverride, err := newOverride(override, logger)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		overrideSlice = append(overrideSlice, validOverride)

	}

	return overrideSlice, errors.Join(errs...)
}
