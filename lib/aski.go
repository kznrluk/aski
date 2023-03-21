package lib

import (
	"bufio"
	"context"
	"fmt"
	"github.com/kznrluk/aski/config"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"os"
)

func getProfile(cfg config.Config, target string) config.Profile {
	for _, profile := range cfg.Profiles {
		if target != "" && profile.ProfileName == target {
			return profile
		}
		if target == "" && profile.Current {
			return profile
		}
	}
	fmt.Printf("WARN: Valid profile not found, using default profile.\n")
	return config.DefaultProfile()
}

func Aski(cmd *cobra.Command, args []string) {
	profileTarget, err := cmd.Flags().GetString("profile")
	isRestMode, _ := cmd.Flags().GetBool("rest")
	content, _ := cmd.Flags().GetString("content")
	system, _ := cmd.Flags().GetString("system")

	checkAPIKey()
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	if err != nil {
		profileTarget = ""
	}

	prof := getProfile(cfg, profileTarget)

	var ctx []openai.ChatCompletionMessage
	if system != "" {
		ctx = append(ctx, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: system,
		})
	} else {
		ctx = append(ctx, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: prof.SystemContext,
		})
	}

	for _, i := range prof.UserMessages {
		ctx = append(ctx, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Name:    prof.UserName,
			Content: i,
		})
	}

	if content != "" {
		ctx = append(ctx, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Name:    prof.UserName,
			Content: content,
		})
		_, _ = Single(cfg, ctx, isRestMode)
	} else {
		StartDialog(cfg, prof, ctx, isRestMode)
	}
}

func checkAPIKey() {
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	if cfg.OpenAIAPIKey == "" {
		fmt.Printf("Please generate an API key from this URL. Currently, the configuration file is saved in plaintext. \nhttps://platform.openai.com/account/api-keys\n")
		fmt.Printf("\t OpenAI API Key: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			text := scanner.Text()
			fmt.Print("Connecting to OpenAI server ... ")
			oc := openai.NewClient(text)
			_, err := oc.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model: openai.GPT3Dot5Turbo,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleSystem,
							Content: "Say Hi!",
						},
					},
				},
			)

			if err != nil {
				fmt.Printf("Erorr: %s", err.Error())
				os.Exit(1)
			}

			fmt.Println("OK")
			cfg.OpenAIAPIKey = text
			if err := config.Save(cfg); err != nil {
				fmt.Printf("Error: %s", err.Error())
				os.Exit(1)
			}
		}
	}
}
