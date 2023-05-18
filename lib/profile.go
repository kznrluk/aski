package lib

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/kznrluk/aski/config"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func ChangeProfile(cmd *cobra.Command, args []string) {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	profileDir := config.MustGetProfileDir()

	var yamlFiles []string

	fileInfo, err := os.ReadDir(profileDir)
	if err != nil {
		panic(err)
	}

	for _, file := range fileInfo {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") && file.Name() != "config.yaml" {
			yamlFiles = append(yamlFiles, file.Name())
		}
	}

	var selected string
	prompt := &survey.Select{
		Message: "Choose one option:",
		Options: yamlFiles,
	}

	_ = survey.AskOne(prompt, &selected)

	cfg.CurrentProfile = selected

	if err := config.Save(cfg); err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
	return
}
