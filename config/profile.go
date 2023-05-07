package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
)

func GetProfile(cfg Config, target string) Profile {
	var p Profile
	found := false
	for _, profile := range cfg.Profiles {
		if target != "" && profile.ProfileName == target {
			found = true
			p = profile
		}
		if target == "" && profile.Current {
			found = true
			p = profile
		}
	}

	if !found {
		fmt.Printf("WARN: Valid profile not found, using default profile.\n")
		initCfg := InitialConfig()
		return initCfg.Profiles[0]
	}

	if err := validateProfile(p); err != nil {
		fmt.Printf("ERROR: Invalid profile: %s\n", err)
		os.Exit(1)
	}
	return p
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
