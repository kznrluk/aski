package config

import (
	_ "embed"
	"fmt"
	"github.com/goccy/go-yaml"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

type Config struct {
	OpenAIAPIKey    string `yaml:"OpenAIAPIKey"`
	AnthropicAPIKey string `yaml:"AnthropicAPIKey"`
	CurrentProfile  string `yaml:"CurrentProfile"`
}

func InitialConfig() Config {
	currentUser, err := user.Current()
	if err != nil {
		currentUser.Username = "aski"
	}
	return Config{
		OpenAIAPIKey:    "",
		AnthropicAPIKey: "",
		CurrentProfile:  GetDefaultProfileFileName(),
	}
}

func GetHomeDir() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	} else if home := os.Getenv("USERPROFILE"); home != "" {
		return home, nil
	} else {
		return "", fmt.Errorf("failed to get home directory, please set $HOME or $USERPROFILE")
	}
}

func MustGetAskiDir() string {
	str, err := GetHomeDir()
	if err != nil {
		fmt.Printf("failed to get aski dir: %s\n", err)
		os.Exit(1)
	}

	return filepath.Join(str, ".aski")
}

func MustGetHistoryDir() string {
	str := MustGetAskiDir()

	return filepath.Join(str, "history")
}

func MustGetProfileDir() string {
	str := MustGetAskiDir()

	return filepath.Join(str, "profile")
}

func CreateInitialConfigFiles() error {
	configDir := MustGetAskiDir()

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	config := InitialConfig()
	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, data, 0600)
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

	configDir := MustGetAskiDir()

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

func GetConfig() (Config, error) {
	askiPath := MustGetAskiDir()

	configPath := filepath.Join(askiPath, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err := CreateInitialConfigFiles()
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

	if config.CurrentProfile == "" {
		config.CurrentProfile = GetDefaultProfileFileName()
		err := Save(config)
		if err != nil {
			return Config{}, fmt.Errorf("failed to save config: %w", err)
		}
	}

	return config, nil
}

func OpenConfigDir() bool {
	var cmd *exec.Cmd

	askiDir := MustGetAskiDir()

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", askiDir)
	case "darwin":
		cmd = exec.Command("open", askiDir)
	case "linux":
		cmd = exec.Command("xdg-open", askiDir)
	default:
		fmt.Printf("unsupported platform: %s \n", runtime.GOOS)
		return false
	}

	err := cmd.Run()
	if err != nil {
		return false
	}

	return true
}
