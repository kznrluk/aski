package lib

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/kznrluk/aski/config"
	"github.com/spf13/cobra"
	"os"
)

func ChangeProfile(cmd *cobra.Command, args []string) {
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	profiles := cfg.Profiles

	strs := []string{}
	for _, p := range profiles {
		strs = append(strs, fmt.Sprintf("%s", p.ProfileName))
	}

	var selected string
	prompt := &survey.Select{
		Message: "Choose one option:",
		Options: strs,
	}

	_ = survey.AskOne(prompt, &selected)

	for i, p := range profiles {
		if selected == p.ProfileName {
			profiles[i].Current = true
		} else {
			profiles[i].Current = false
		}
	}

	cfg.Profiles = profiles

	if err := config.Save(cfg); err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
	return
}
