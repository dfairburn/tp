package config

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"

	logging "github.com/sirupsen/logrus"
)

func LoadVars(logger *logging.Logger, paths ...string) map[interface{}]interface{} {
	y := make(map[interface{}]interface{})

	// this gives precedent to paths passed in via config and flags, then processes the default file paths
	paths = append(paths, buildVarPaths().paths...)

	f, err := tryFiles(logger, paths...)
	if err != nil {
		logger.Errorf("unable to use var files, error: %v", err)
		return y
	}

	data, err := io.ReadAll(f)
	if err != nil {
		logger.Fatalf("error: %v", err)
	}

	err = yaml.Unmarshal(data, &y)
	if err != nil {
		logger.Fatalf("error: %v", err)
	}

	return y
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
		logger.Warnf("extra colon used in override. If these are meant to be separaate overrides, please add"+
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
