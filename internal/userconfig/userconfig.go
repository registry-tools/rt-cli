package config

import (
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type UserConfiguration interface {
	SaveToUserConfigDirectory(version string) error
	SetHostToken(hostname, token string)
	GetHostToken(hostname string) (string, bool)
}

type hostConfig struct {
	Hostname string `yaml:"hostname"`
	Token    string `yaml:"token"`
}

type userConfig struct {
	filePath         string
	Hosts            []*hostConfig `yaml:"hosts"`
	CreatedByVersion string        `yaml:"created_by_version"`
}

func (c *userConfig) GetHostToken(hostname string) (string, bool) {
	for _, host := range c.Hosts {
		if host.Hostname == hostname {
			return host.Token, true
		}
	}

	return "", false
}

func (c *userConfig) SetHostToken(hostname, token string) {
	for _, host := range c.Hosts {
		if host.Hostname == hostname {
			host.Token = token
			return
		}
	}

	c.Hosts = append(c.Hosts, &hostConfig{
		Hostname: hostname,
		Token:    token,
	})
}

func (c *userConfig) SaveToUserConfigDirectory(cliVersion string) error {
	err := os.MkdirAll(path.Dir(c.filePath), 0700)
	if err != nil {
		return fmt.Errorf("error creating user config directory %q: %w", c.filePath, err)
	}

	c.CreatedByVersion = cliVersion

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error encoding user config: %w", err)
	}

	err = os.WriteFile(c.filePath, data, 0600)
	if err != nil {
		return fmt.Errorf("error writing user config: %w", err)
	}

	return nil
}

func LoadFromUserConfigDirectory() (UserConfiguration, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("error reading user config: %w", err)
	}

	configPath := path.Join(configDir, "registrytools", "config.yaml")

	var result = userConfig{
		Hosts:    make([]*hostConfig, 0),
		filePath: configPath,
	}

	info, err := os.Stat(configPath)
	if err != nil && os.IsNotExist(err) {
		// Initialize an empty config if it doesn't exist
		return &result, nil
	} else if err != nil {
		return nil, fmt.Errorf("error reading user config: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("error loading user config: %s is a directory", configPath)
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading user config: %w", err)
	}

	if err = yaml.Unmarshal(configData, &result); err != nil {
		return nil, fmt.Errorf("error decoding user config: %w", err)
	}

	return &result, nil
}
