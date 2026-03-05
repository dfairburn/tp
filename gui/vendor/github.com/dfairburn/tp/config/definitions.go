package config

import (
	"strings"

	"github.com/dfairburn/tp/paths"
)

const (
	ConfigDir      = "config/"
	ConfigFilename = "config"

	YamlExt = ".yaml"
	YmlExt  = ".yml"
	EnvExt  = ".env"

	LogFilename = "tp.log"

	DefaultEnvFilename = "default" + EnvExt

	HomeLoc = "~/.tp/"
	RelLoc  = "./tp/"
	HereLoc = "./"
)

var (
	DefaultTemplatesDirectory = paths.Expand(HomeLoc + "templates")
	DefaultDirectory          = paths.Expand(HomeLoc)
	DefaultLogFile            = paths.Expand(HomeLoc + LogFilename)
	DefaultConfigFile         = paths.Expand(HomeLoc + ConfigFilename + YmlExt)
	DefaultEnvPath            = paths.Expand(HomeLoc + DefaultEnvFilename + YmlExt)
	DefaultEnv                = "default"
)

type pathRepo struct {
	paths []string
}

func (p *pathRepo) add(path ...string) {
	p.paths = append(p.paths, strings.Join(path, ""))
}

func buildConfigPaths() pathRepo {
	p := pathRepo{}
	p.add(HomeLoc, ConfigDir, ConfigFilename, YamlExt)
	p.add(HomeLoc, ConfigDir, ConfigFilename, YmlExt)

	p.add(HomeLoc, ConfigFilename, YamlExt)
	p.add(HomeLoc, ConfigFilename, YmlExt)

	p.add(RelLoc, ConfigDir, ConfigFilename, YamlExt)
	p.add(RelLoc, ConfigDir, ConfigFilename, YmlExt)

	p.add(RelLoc, ConfigFilename, YamlExt)
	p.add(RelLoc, ConfigFilename, YmlExt)

	p.add(HereLoc, ConfigFilename, YamlExt)
	p.add(HereLoc, ConfigFilename, YmlExt)

	return p
}

func buildEnvPaths() pathRepo {
	p := pathRepo{}

	p.add(HomeLoc, DefaultEnvFilename, YamlExt)
	p.add(HomeLoc, DefaultEnvFilename, YmlExt)

	p.add(HomeLoc, YamlExt)
	p.add(HomeLoc, YmlExt)

	p.add(RelLoc, DefaultEnvFilename, YamlExt)
	p.add(RelLoc, DefaultEnvFilename, YmlExt)

	p.add(RelLoc, DefaultEnvFilename, YamlExt)
	p.add(RelLoc, DefaultEnvFilename, YmlExt)

	p.add(HereLoc, DefaultEnvFilename, YamlExt)
	p.add(HereLoc, DefaultEnvFilename, YmlExt)

	return p
}
