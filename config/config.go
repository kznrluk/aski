package config

import (
	_ "embed"
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/sashabaranov/go-openai"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

type Profile struct {
	ProfileName   string   `yaml:"ProfileName"`
	UserName      string   `yaml:"UserName"`
	Current       bool     `yaml:"Current"`
	SystemContext string   `yaml:"SystemContext"`
	UserMessages  []string `yaml:"UserMessages"`
	Model         string   `yaml:"Model"`
}

type Config struct {
	OpenAIAPIKey string    `yaml:"OpenAIAPIKey"`
	AutoSave     bool      `yaml:"AutoSave"`
	Summarize    bool      `yaml:"Summarize"`
	Profiles     []Profile `yaml:"Profiles"`
}

func InitialConfig() Config {
	currentUser, err := user.Current()
	if err != nil {
		currentUser.Username = "aski"
	}
	return Config{
		OpenAIAPIKey: "",
		AutoSave:     true,
		Summarize:    true,
		Profiles: []Profile{
			{
				ProfileName:   "GPT3.5",
				UserName:      currentUser.Username,
				Current:       true,
				SystemContext: "You are a kind and helpful chat AI. Sometimes you may say things that are incorrect, but that is unavoidable.",
				Model:         openai.GPT3Dot5Turbo,
				UserMessages:  []string{},
			},
			{
				ProfileName:   "GPT3.5Emoji",
				UserName:      currentUser.Username,
				Current:       false,
				SystemContext: "You are a kind and helpful chat AI. Sometimes you may say things that are incorrect, but that is unavoidable. With lot of emojis.",
				Model:         openai.GPT3Dot5Turbo,
				UserMessages:  []string{},
			},
		},
	}
}

func GetHomeDir() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	} else if home := os.Getenv("USERPROFILE"); home != "" {
		return home, nil
	} else {
		return "", fmt.Errorf("cannot find home directory")
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

	data, err := yaml.Marshal(InitialConfig())
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
	homeDir, err := GetHomeDir()
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

func OpenConfigDir() bool {
	var cmd *exec.Cmd

	home, err := GetHomeDir()
	if err != nil {
		fmt.Printf("can't get home dir")
	}

	aski := home + "/.aski"

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", aski)
	case "darwin":
		cmd = exec.Command("open", aski)
	case "linux":
		cmd = exec.Command("xdg-open", aski)
	default:
		fmt.Printf("unsupported platform: %s \n", runtime.GOOS)
		return false
	}

	err = cmd.Run()
	if err != nil {
		return false
	}

	return true
}
