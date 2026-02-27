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

func LoadEnvironment(logger *logging.Logger, paths ...string) (string, map[interface{}]interface{}) {
	y := make(map[interface{}]interface{})

	// this gives precedent to paths passed in via config and flags, then processes the default file paths
	paths = append(paths, buildEnvPaths().paths...)

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

	return path, y
}

// ExpandVarsForKeys selectively expands only the variables whose keys are in the needed set.
// For nested maps, it recursively descends and expands matching keys at any depth.
// Variables not in the needed set are copied through with their raw (unexpanded) values.
func ExpandVarsForKeys(logger *logging.Logger, y map[any]any, needed map[string]bool) (map[any]any, error) {
	expandedMap := make(map[any]any)

	for key, value := range y {
		keyStr, ok := key.(string)
		if !ok {
			expandedMap[key] = value
			continue
		}

		switch v := value.(type) {
		case string:
			if needed[keyStr] {
				expanded, didExpand, err := expandVariables(logger, keyStr, v)
				if err != nil {
					return nil, err
				}
				if didExpand {
					expandedMap[key] = expanded
				} else {
					expandedMap[key] = value
				}
			} else {
				expandedMap[key] = value
			}
		case map[string]interface{}:
			// cast map[string]interface{} to map[any]any for recursive call
			m := make(map[interface{}]interface{})
			for k, vv := range v {
				m[k] = vv
			}

			expanded, err := ExpandVarsForKeys(logger, m, needed)
			if err != nil {
				return nil, err
			}
			expandedMap[key] = expanded
		default:
			expandedMap[key] = value
		}
	}

	return expandedMap, nil
}

func expandVariables(logger *logging.Logger, key, value string) (string, bool, error) {
	v := value
	re := regexp.MustCompile("\\$\\((?P<command>.*)\\)")
	result := make(map[string]string)
	if !re.MatchString(v) {
		return "", false, nil
	}

	match := re.FindStringSubmatch(v)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	cmd, ok := result["command"]
	if !ok {
		return "", false, nil
	}

	shell := os.Getenv("SHELL")
	e := exec.Command(shell, "-c", cmd)
	var out strings.Builder
	var outErr strings.Builder
	e.Stdout = &out
	e.Stderr = &outErr
	err := e.Run()
	if err != nil {
		return "", false, fmt.Errorf("error expanding variable %q, executing command %q: %w", key, cmd, err)
	}

	expanded := strings.TrimSuffix(out.String(), "\n")
	return expanded, true, nil
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
