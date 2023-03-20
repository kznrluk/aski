package config

import (
	_ "embed"
	"fmt"
	"github.com/goccy/go-yaml"
	"io"
	"os"
	"path/filepath"
)

//go:embed config.yaml.example
var DEFAULT_CONFIG []byte

type Profile struct {
	ProfileName   string   `yaml:"ProfileName"`
	UserName      string   `yaml:"UserName"`
	Current       bool     `yaml:"Current"`
	SystemContext string   `yaml:"SystemContext"`
	UserMessages  []string `yaml:"UserMessages"`
}

type Config struct {
	OpenAIAPIKey string    `yaml:"OpenAIAPIKey"`
	Profiles     []Profile `yaml:"Profiles"`
}

func getHomeDir() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	} else if home := os.Getenv("USERPROFILE"); home != "" {
		return home, nil
	} else {
		return "", fmt.Errorf("cannot find home directory")
	}
}

func DefaultProfile() Profile {
	return Profile{
		ProfileName:   "Default",
		UserName:      "AskiUser",
		Current:       true,
		SystemContext: "You are a kind and helpful chat AI. Sometimes you may say things that are incorrect, but that is unavoidable.",
		UserMessages:  []string{},
	}
}

func InitSave() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".aski")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, DEFAULT_CONFIG, 0600)
	if err != nil {
		return err
	}
	return nil
}

func Save(config Config) error {
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".aski")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, yamlData, 0600)
	if err != nil {
		return err
	}
	return nil
}

func Init() (Config, error) {
	homeDir, err := getHomeDir()
	if err != nil {
		return Config{}, err
	}

	configPath := homeDir + "/.aski/config.yaml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err := InitSave()
		if err != nil {
			return Config{}, err
		}
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		return Config{}, err
	}

	defer configFile.Close()

	configBytes, err := io.ReadAll(configFile)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
