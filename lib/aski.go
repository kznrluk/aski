package lib

import (
	"bufio"
	"context"
	"fmt"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/ctx"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

type FileContents struct {
	Name     string
	Path     string
	Contents string
	Length   int
}

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

func isBinary(contents []byte) bool {
	for _, ch := range contents {
		if ch == 0 {
			return true
		}
	}
	return false
}

func Aski(cmd *cobra.Command, args []string) {
	profileTarget, err := cmd.Flags().GetString("profile")
	isRestMode, _ := cmd.Flags().GetBool("rest")
	content, _ := cmd.Flags().GetString("content")
	fileGlobs, _ := cmd.Flags().GetStringSlice("file")

	checkAPIKey()
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	if err != nil {
		profileTarget = ""
	}

	prof := getProfile(cfg, profileTarget)
	ctx := ctx.NewContext(prof, cfg)
	ctx.Append(openai.ChatMessageRoleSystem, prof.SystemContext)

	if len(fileGlobs) != 0 {
		fileContents := getFileContents(fileGlobs)
		for _, f := range fileContents {
			if content == "" {
				fmt.Printf("Append File: %s\n", f.Name)
			}
			ctx.Append(openai.ChatMessageRoleUser, fmt.Sprintf("Path: `%s`\n ```\n%s```", f.Path, f.Contents))
		}
	}

	for _, i := range prof.UserMessages {
		ctx.Append(openai.ChatMessageRoleUser, i)
	}

	if content != "" {
		ctx.Append(openai.ChatMessageRoleUser, content)
		_, _ = Single(cfg, prof, ctx, isRestMode)
	} else {
		StartDialog(cfg, prof, ctx, isRestMode)
	}
}

func getFileContents(fileGlobs []string) []FileContents {
	var fileContents []FileContents
	for _, arg := range fileGlobs {
		files, err := filepath.Glob(arg)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			contentsBytes, err := os.ReadFile(file)
			if err != nil {
				panic(err)
			}
			content := string(contentsBytes)
			if isBinary(contentsBytes) {
				continue
			}

			info, err := os.Stat(file)
			if err != nil {
				panic(err)
			}

			fileContents = append(fileContents, FileContents{
				Name:     info.Name(),
				Path:     file,
				Contents: content,
				Length:   len(content),
			})
		}
	}
	return fileContents
}

func checkAPIKey() {
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	if cfg.OpenAIAPIKey == "" {
		fmt.Printf("Please generate an API key from this URL. Headly, the configuration file is saved in plaintext. \nhttps://platform.openai.com/account/api-keys\n")
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