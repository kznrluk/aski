package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/kznrluk/aski/config"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func getProfile(cfg config.Config, target string) config.Profile {
	for _, profile := range cfg.Profiles {
		if target != "" && profile.ProfileName == target {
			return profile
		} else if target == "" && profile.Current {
			return profile
		}
	}
	fmt.Printf("WARN: Valid profile not found, using default profile.\n")
	return config.DefaultProfile()
}

func main() {
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	var rootCmd = &cobra.Command{
		Use:   "aski",
		Short: "aski is a very small and user-friendly ChatGPT client.",
		Long:  `aski is a very small and user-friendly ChatGPT client. It works hard to maintain context and establish communication.`,
		Run: func(cmd *cobra.Command, args []string) {
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

			oc := openai.NewClient(cfg.OpenAIAPIKey)

			profileTarget, err := cmd.Flags().GetString("profile")
			if err != nil {
				profileTarget = ""
			}

			prof := getProfile(cfg, profileTarget)

			var ctx []openai.ChatCompletionMessage
			ctx = append(ctx, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: prof.SystemContext,
			})

			for _, i := range prof.UserMessages {
				ctx = append(ctx, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Name:    prof.UserName,
					Content: i,
				})
			}

			isRestMode, _ := cmd.Flags().GetBool("rest")
			if isRestMode {
				fmt.Printf("REST Mode \n")
			}

			scanner := bufio.NewScanner(os.Stdin)
			fmt.Printf("%s@%s> ", prof.UserName, prof.ProfileName)
			for scanner.Scan() {
				yellow := color.New(color.FgHiYellow).SprintFunc()

				input := scanner.Text()

				ctx = append(ctx, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Name:    prof.UserName,
					Content: input,
				})

				data := ""
				if !isRestMode {
					stream, err := oc.CreateChatCompletionStream(
						context.Background(),
						openai.ChatCompletionRequest{
							Model:    openai.GPT3Dot5Turbo,
							Messages: ctx,
						},
					)

					if err != nil {
						fmt.Printf(err.Error())
						continue
					}

					for {
						resp, err := stream.Recv()
						if err != nil {
							if err == io.EOF {
								break
							} else {
								fmt.Printf(err.Error())
								continue
							}
						}

						fmt.Printf("%s", yellow(resp.Choices[0].Delta.Content))
						data += resp.Choices[0].Delta.Content
					}
				} else {
					resp, err := oc.CreateChatCompletion(
						context.Background(),
						openai.ChatCompletionRequest{
							Model:    openai.GPT3Dot5Turbo,
							Messages: ctx,
						},
					)

					if err != nil {
						fmt.Printf(err.Error())
						continue
					}
					fmt.Printf("%s", yellow(resp.Choices[0].Message.Content))
					data = resp.Choices[0].Message.Content
				}

				ctx = append(ctx, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: data,
				})

				fmt.Printf("\n\n%s@%s> ", prof.UserName, prof.ProfileName)
			}
		},
	}

	changeProfileCmd := &cobra.Command{
		Use:   "change-profile",
		Short: "Change profile.",
		Long: "Profiles are usually located in the .aski/config.yaml file in the home directory." +
			"By using profiles, you can easily switch between different conversation contexts on the fly.",
		Run: func(cmd *cobra.Command, args []string) {
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

			for _, p := range profiles {
				if selected == p.ProfileName {
					p.Current = true
				} else {
					p.Current = false
				}
			}

			if err := config.Save(cfg); err != nil {
				fmt.Printf("Error: %s", err.Error())
				os.Exit(1)
			}
			return
		},
	}

	rootCmd.AddCommand(changeProfileCmd)
	rootCmd.PersistentFlags().StringP("profile", "p", "", "Select the profile to use for this conversation, as defined in the .aski/config.yaml file.")
	rootCmd.PersistentFlags().BoolP("rest", "r", false, "When you specify this flag, you will communicate with the REST API instead of streaming. This can be useful if the communication is unstable or if you are not receiving responses properly.")

	rootCmd.Execute()
}
