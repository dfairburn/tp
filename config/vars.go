package config

import (
	"gopkg.in/yaml.v3"
	"io"

	logging "github.com/sirupsen/logrus"
)

const (
	VarHomeLocInDirFile = HomeLoc + VarDir + VarFilename
	VarHomeLocFile      = HomeLoc + VarFilename
	VarRelLocInDirFile  = RelLoc + VarDir + VarFilename
	VarRelLocFile       = RelLoc + VarFilename
)

var (
	varPathsToCheck = []string{
		VarHomeLocInDirFile,
		VarHomeLocFile,
		VarRelLocInDirFile,
		VarRelLocFile,
	}
)

func LoadVars(logger *logging.Logger, paths ...string) map[interface{}]interface{} {
	y := make(map[interface{}]interface{})

	// this gives precedent to paths passed in via config and flags, then processes the default file paths
	paths = append(paths, varPathsToCheck...)

	f, err := tryFiles(logger, paths...)
	if err != nil {
		logger.Fatalf("error: %v", err)
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
