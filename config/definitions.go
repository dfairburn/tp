package config

import "strings"

const (
	ConfigDir      = "config/"
	ConfigFilename = "config"

	YamlExt = ".yaml"
	YmlExt  = ".yml"

	LogFilename = "tp.log"

	VarDir      = "vars/"
	VarFilename = "vars"

	HomeLoc = "~/.tp/"
	RelLoc  = "./tp/"
	HereLoc = "./"

	DefaultLogFile    = HomeLoc + LogFilename
	DefaultConfigFile = HomeLoc + ConfigFilename + YamlExt
	DefaultVarFile    = HomeLoc + VarFilename + YamlExt
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

func buildVarPaths() pathRepo {
	p := pathRepo{}

	p.add(HomeLoc, VarDir, VarFilename, YamlExt)
	p.add(HomeLoc, VarDir, VarFilename, YmlExt)

	p.add(HomeLoc, VarFilename, YamlExt)
	p.add(HomeLoc, VarFilename, YmlExt)

	p.add(RelLoc, VarDir, VarFilename, YamlExt)
	p.add(RelLoc, VarDir, VarFilename, YmlExt)

	p.add(RelLoc, VarFilename, YamlExt)
	p.add(RelLoc, VarFilename, YmlExt)

	p.add(HereLoc, VarFilename, YamlExt)
	p.add(HereLoc, VarFilename, YmlExt)

	return p
}
