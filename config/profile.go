package config

import (
	"errors"
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/sashabaranov/go-openai"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

type Profile struct {
	ProfileName      string           `yaml:"ProfileName"`
	Model            string           `yaml:"Model"`
	UserName         string           `yaml:"UserName"`
	AutoSave         bool             `yaml:"AutoSave"`
	Summarize        bool             `yaml:"Summarize"`
	SystemContext    string           `yaml:"SystemContext"`
	Messages         []PreMessage     `yaml:"Messages"`
	CustomParameters CustomParameters `yaml:"CustomParameters,omitempty"`
}

type PreMessage struct {
	Role    string `yaml:"Role"`
	Content string `yaml:"Content"`
}

// CustomParameters - When these parameters are specified, they will be overwritten during transmission.
type CustomParameters struct {
	MaxTokens        int            `yaml:"max_tokens,omitempty"`
	Temperature      float32        `yaml:"temperature,omitempty"`
	TopP             float32        `yaml:"top_p,omitempty"`
	Stop             []string       `yaml:"stop,omitempty"`
	PresencePenalty  float32        `yaml:"presence_penalty,omitempty"`
	FrequencyPenalty float32        `yaml:"frequency_penalty,omitempty"`
	LogitBias        map[string]int `yaml:"logit_bias,omitempty"`
	// N is fixed at 1 currently
	// N                int            `yaml:"n,omitempty"`
}

func GetDefaultProfileFileName() string {
	return "default.yaml"
}

func GetProfile(cfg Config, overload string) (Profile, error) {
	if !hasDefaultProfile() {
		err := CreateInitialProfileFile()
		if err != nil {
			return Profile{}, fmt.Errorf("cannot create initial profile file: %s", err)
		}
	}
	// We called hasDefaultProfile() above, so we know that the default profile exists
	profileDir := MustGetProfileDir()
	toSearchPaths := createToSearchPaths(profileDir, cfg, overload)

	for _, target := range toSearchPaths {
		if _, err := os.Stat(target); os.IsNotExist(err) {
			continue
		}

		profileFile, err := os.Open(target)
		if err != nil {
			return Profile{}, fmt.Errorf("cannot open profile file: %s", err)
		}
		defer profileFile.Close()

		profileBytes, err := io.ReadAll(profileFile)
		if err != nil {
			return Profile{}, fmt.Errorf("cannot read profile file: %s", err)
		}

		var profile Profile
		err = yaml.Unmarshal(profileBytes, &profile)
		if err != nil {
			return Profile{}, fmt.Errorf("cannot parse profile file: %s", err)
		}

		// Validate the loaded profile
		if err := validateProfile(profile); err != nil {
			return Profile{}, fmt.Errorf("invalid profile %s: %s", target, err)
		}

		return profile, nil
	}

	return Profile{}, fmt.Errorf("profile file not found, tried: %s", strings.Join(toSearchPaths, ", "))
}

func createToSearchPaths(profileDir string, cfg Config, overload string) []string {
	toSearchPaths := []string{}
	if overload == "" {
		// Use the profile specified in the config file
		toSearchPaths = append(toSearchPaths, filepath.Join(profileDir, cfg.CurrentProfile))
	} else {
		// Current directory
		toSearchPaths = append(toSearchPaths, overload)

		// Profile directory
		if !(strings.Contains(overload, "/") || strings.Contains(overload, "\\")) {
			if strings.HasSuffix(overload, ".yaml") || strings.HasSuffix(overload, ".yml") {
				toSearchPaths = append(toSearchPaths, filepath.Join(profileDir, overload))
			} else {
				toSearchPaths = append(toSearchPaths, filepath.Join(profileDir, overload+".yaml"))
				toSearchPaths = append(toSearchPaths, filepath.Join(profileDir, overload+".yml"))
			}
		}
	}
	return toSearchPaths
}

func hasDefaultProfile() bool {
	profileDir := MustGetProfileDir()
	defaultProfilePath := filepath.Join(profileDir, "default.yaml")
	if _, err := os.Stat(defaultProfilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func CreateInitialProfileFile() error {
	askiDir := MustGetAskiDir()

	profileDir := filepath.Join(askiDir, "profile")
	if err := os.MkdirAll(profileDir, 0700); err != nil {
		return err
	}

	profileData, err := yaml.Marshal(InitialProfile())
	if err != nil {
		return err
	}

	profilePath := filepath.Join(profileDir, GetDefaultProfileFileName())
	err = os.WriteFile(profilePath, profileData, 0700)
	if err != nil {
		return err
	}

	return nil
}

func InitialProfile() Profile {
	currentUser, err := user.Current()
	if err != nil {
		currentUser.Username = "aski"
	}
	return Profile{
		ProfileName:   "GPT3.5",
		UserName:      currentUser.Username,
		AutoSave:      true,
		Summarize:     true,
		SystemContext: "You are a kind and helpful chat AI. Sometimes you may say things that are incorrect, but that is unavoidable.",
		Model:         openai.GPT3Dot5Turbo,
		Messages:      []PreMessage{},
	}
}

func validateProfile(profile Profile) error {
	if profile.ProfileName == "" {
		return fmt.Errorf("ProfileName must not be empty")
	}
	if len(profile.UserName) > 16 || !regexp.MustCompile("^[a-zA-Z0-9]+$").MatchString(profile.UserName) {
		return fmt.Errorf("UserName must be alphanumeric and no more than 8 characters")
	}
	if profile.SystemContext == "" {
		return fmt.Errorf("SystemContext must not be empty")
	}
	if profile.Model == "" {
		return fmt.Errorf("Model must not be empty")
	}

	for _, message := range profile.Messages {
		if message.Role == "" {
			return fmt.Errorf("Message Role must not be empty")
		}
		if message.Content == "" {
			return fmt.Errorf("Message Content must not be empty")
		}
	}

	return ValidateCustomParameters(profile.CustomParameters)
}

func ValidateCustomParameters(customParams CustomParameters) error {
	if customParams.Temperature != 0 && (customParams.Temperature < 0 || customParams.Temperature > 2) {
		return errors.New("temperature must be between 0 and 2")
	}
	if customParams.TopP != 0 && (customParams.TopP < 0 || customParams.TopP > 1) {
		return errors.New("top_p must be between 0 and 1")
	}
	if len(customParams.Stop) > 4 {
		return errors.New("stop can contain up to 4 sequences")
	}
	if customParams.PresencePenalty != 0 && (customParams.PresencePenalty < -2 || customParams.PresencePenalty > 2) {
		return errors.New("presence_penalty must be between -2 and 2")
	}
	if customParams.FrequencyPenalty != 0 && (customParams.FrequencyPenalty < -2 || customParams.FrequencyPenalty > 2) {
		return errors.New("frequency_penalty must be between -2 and 2")
	}
	for _, bias := range customParams.LogitBias {
		if bias != 0 && (bias < -100 || bias > 100) {
			return errors.New("logit_bias values must be between -100 and 100")
		}
	}
	return nil
}
