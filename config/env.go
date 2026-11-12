package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	tppaths "github.com/dfairburn/tp/paths"
	"gopkg.in/yaml.v3"

	logging "github.com/sirupsen/logrus"
)

func ResolveEnvironmentFile(logger *logging.Logger, configuredPath, environmentName, explicitPath string) (string, error) {
	if explicitPath != "" {
		return tppaths.Expand(explicitPath), nil
	}

	if environmentName != "" {
		canonicalPath := environmentFilePath(environmentName)
		legacyPath := legacyEnvironmentFilePath(environmentName)
		return migrateLegacyEnvironmentFile(logger, canonicalPath, legacyPath)
	}

	if configuredPath == "" {
		configuredPath = DefaultEnvironmentFilePath()
	}

	canonicalPath := tppaths.Expand(configuredPath)
	defaultPath := DefaultEnvironmentFilePath()
	if filepath.Clean(canonicalPath) == filepath.Clean(defaultPath) {
		return migrateLegacyEnvironmentFile(logger, canonicalPath, legacyEnvironmentFilePath(DefaultEnv))
	}

	return canonicalPath, nil
}

func environmentFilePath(name string) string {
	return tppaths.Expand(HomeLoc + name + EnvExt + YmlExt)
}

func legacyEnvironmentFilePath(name string) string {
	return tppaths.Expand(HomeLoc + name + YmlExt)
}

func migrateLegacyEnvironmentFile(logger *logging.Logger, canonicalPath, legacyPath string) (string, error) {
	_, canonicalErr := os.Stat(canonicalPath)
	if canonicalErr == nil {
		if _, legacyErr := os.Stat(legacyPath); legacyErr == nil {
			logger.Warnf("both canonical environment file %q and legacy file %q exist; using the canonical file", canonicalPath, legacyPath)
		} else if !errors.Is(legacyErr, os.ErrNotExist) {
			return "", fmt.Errorf("checking legacy environment file %q: %w", legacyPath, legacyErr)
		}
		return canonicalPath, nil
	}
	if !errors.Is(canonicalErr, os.ErrNotExist) {
		return "", fmt.Errorf("checking canonical environment file %q: %w", canonicalPath, canonicalErr)
	}

	if _, legacyErr := os.Stat(legacyPath); legacyErr != nil {
		if errors.Is(legacyErr, os.ErrNotExist) {
			return canonicalPath, nil
		}
		return "", fmt.Errorf("checking legacy environment file %q: %w", legacyPath, legacyErr)
	}

	if err := os.Link(legacyPath, canonicalPath); err != nil {
		if errors.Is(err, os.ErrExist) {
			logger.Warnf("canonical environment file %q appeared during migration; using it and leaving legacy file %q untouched", canonicalPath, legacyPath)
			return canonicalPath, nil
		}
		logger.Warnf("could not migrate legacy environment file %q to %q: %v", legacyPath, canonicalPath, err)
		return "", fmt.Errorf("migrating legacy environment file %q to %q: %w", legacyPath, canonicalPath, err)
	}
	if err := os.Remove(legacyPath); err != nil {
		logger.Warnf("migrated environment file to %q but could not remove legacy path %q: %v", canonicalPath, legacyPath, err)
		return "", fmt.Errorf("removing legacy environment file %q after migration: %w", legacyPath, err)
	}

	logger.Infof("migrated legacy environment file %q to %q", legacyPath, canonicalPath)
	return canonicalPath, nil
}

func LoadEnvironment(logger *logging.Logger, environmentPath string) (string, map[interface{}]interface{}, error) {
	y := make(map[interface{}]interface{})
	path := tppaths.Expand(environmentPath)
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warnf("environment file %q does not exist; continuing without file variables", path)
			return path, y, nil
		}
		return path, nil, fmt.Errorf("opening environment file %q: %w", path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return path, nil, fmt.Errorf("reading environment file %q: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &y); err != nil {
		return path, nil, fmt.Errorf("parsing environment file %q: %w", path, err)
	}

	variables := expandVars(logger, y)
	logger.Debugf("loaded environment file %q", path)
	return path, variables, nil
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
