package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/paths"
	"gopkg.in/yaml.v3"
)

// GetConfig returns the current configuration
func (a *App) GetConfig() ConfigInfo {
	return ConfigInfo{
		ConfigPath:             a.configPath,
		TemplatesDir:           a.templatesDir,
		TemplatesDirectoryPath: a.config.TemplatesDirectoryPath,
		EnvironmentFile:        a.config.EnvironmentFile,
	}
}

// SaveConfig writes updated configuration fields to the config file
func (a *App) SaveConfig(environmentFile, templatesDirectoryPath string) error {
	configPath := a.configPath
	if configPath == "" {
		configPath = paths.Expand("~/.tp/config.yml")
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}
	cfg := config.Config{
		EnvironmentFile:        environmentFile,
		TemplatesDirectoryPath: templatesDirectoryPath,
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}
	a.configPath = configPath
	a.config = cfg
	a.templatesDir = paths.Expand(templatesDirectoryPath)
	return nil
}

// ReloadConfig reloads the configuration and environment
func (a *App) ReloadConfig() error {
	cfg, configPath, err := config.LoadOrDefaultConfig(a.logger)
	if err != nil {
		return err
	}
	a.config = cfg
	a.configPath = configPath
	a.templatesDir = paths.Expand(cfg.TemplatesDirectoryPath)
	if a.templatesDir == "" {
		a.templatesDir = paths.Expand("~/.tp/templates")
	}
	_, a.vars = config.LoadEnvironment(a.logger, cfg.EnvironmentFile)
	return nil
}

// GetVariables returns the current environment variables
func (a *App) GetVariables() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range a.vars {
		if key, ok := k.(string); ok {
			result[key] = v
		}
	}
	return result
}

// RefreshVariables reloads environment variables, re-executing any $(command) expansions.
// Useful for refreshing tokens fetched via shell commands.
func (a *App) RefreshVariables() (map[string]interface{}, error) {
	envPath, vars := config.LoadEnvironment(a.logger, a.config.EnvironmentFile)
	a.vars = vars
	a.logger.Infof("RefreshVariables: loaded %d variables from %s", len(vars), envPath)
	result := a.GetVariables()
	a.logger.Infof("RefreshVariables: returning %d variables", len(result))
	return result, nil
}

// GetEnvironmentFilePath returns the path to the environment/variables file
func (a *App) GetEnvironmentFilePath() string {
	envFile := a.config.EnvironmentFile
	if envFile == "" {
		envFile = config.DefaultEnvPath
	}
	return paths.Expand(envFile)
}

// OpenEnvironmentFile opens the environment/variables file in the default editor
func (a *App) OpenEnvironmentFile() error {
	return a.OpenInEditor(a.GetEnvironmentFilePath())
}

// GetOverridesFilePath returns the path to the overrides file
func (a *App) GetOverridesFilePath() string {
	return paths.Expand("~/.tp/overrides.yaml")
}

// GetOverrides loads and returns the current overrides from file
func (a *App) GetOverrides() (map[string]string, error) {
	overridesPath := a.GetOverridesFilePath()
	result := make(map[string]string)

	content, err := os.ReadFile(overridesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}

	var rawOverrides map[string]interface{}
	if err := yaml.Unmarshal(content, &rawOverrides); err != nil {
		return nil, err
	}

	for k, v := range rawOverrides {
		if strVal, ok := v.(string); ok {
			result[k] = strVal
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}

// SaveOverrides saves overrides to the overrides file
func (a *App) SaveOverrides(overrides map[string]string) error {
	overridesPath := a.GetOverridesFilePath()
	if err := os.MkdirAll(filepath.Dir(overridesPath), 0755); err != nil {
		return err
	}
	content, err := yaml.Marshal(overrides)
	if err != nil {
		return err
	}
	return os.WriteFile(overridesPath, content, 0644)
}

// OpenOverridesFile opens the overrides file in the default editor, creating it if needed
func (a *App) OpenOverridesFile() error {
	overridesPath := a.GetOverridesFilePath()
	if _, err := os.Stat(overridesPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(overridesPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(overridesPath, []byte("# Variable overrides\n# Example: token: my-token-value\n"), 0644); err != nil {
			return err
		}
	}
	return a.OpenInEditor(overridesPath)
}
