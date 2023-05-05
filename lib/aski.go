package lib

import (
	"bufio"
	"context"
	"fmt"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/conv"
	"github.com/kznrluk/aski/session"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	initCfg := config.InitialConfig()
	return initCfg.Profiles[0]
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
	restore, _ := cmd.Flags().GetString("restore")
	verbose, _ := cmd.Flags().GetBool("verbose")
	session.SetVerbose(verbose)

	fileInfo, _ := os.Stdin.Stat()
	if (fileInfo.Mode() & os.ModeNamedPipe) != 0 {
		session.SetIsPipe(true)
	}

	checkAPIKey()
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	if err != nil {
		profileTarget = ""
	}

	prof := getProfile(cfg, profileTarget)

	var ctx conv.Conversation
	if restore != "" {
		load, fileName, err := ReadFileFromPWDAndHistoryDir(restore)
		if err != nil {
			fmt.Printf("error reading restore file: %v\n", err)
			os.Exit(1)
		}

		ctx, err = conv.FromYAML(load)
		if err != nil {
			fmt.Printf("error parsing restore file: %v\n", err)
			os.Exit(1)
		}

		if len(fileGlobs) != 0 {
			// TODO: We should be able to renew file contents from the globs
			fmt.Printf("WARN: File globs are ignored when loading restore.\n")
		}

		println("Restore conversations from " + fileName)
	} else {
		ctx = conv.NewConversation(prof)
		ctx.Append(openai.ChatMessageRoleSystem, prof.SystemContext)

		if len(fileGlobs) != 0 {
			fileContents := getFileContents(fileGlobs)
			for _, f := range fileContents {
				if content == "" && !session.IsPipe() {
					fmt.Printf("Append File: %s\n", f.Name)
				}
				ctx.Append(openai.ChatMessageRoleUser, fmt.Sprintf("Path: `%s`\n ```\n%s```", f.Path, f.Contents))
			}
		}

		for _, i := range prof.Messages {
			switch strings.ToLower(i.Role) {
			case openai.ChatMessageRoleSystem:
				ctx.Append(openai.ChatMessageRoleSystem, i.Content)
			case openai.ChatMessageRoleUser:
				ctx.Append(openai.ChatMessageRoleUser, i.Content)
			case openai.ChatMessageRoleAssistant:
				ctx.Append(openai.ChatMessageRoleAssistant, i.Content)
			default:
				panic(fmt.Errorf("invalid role: %s", i.Role))
			}
		}
	}

	if session.IsPipe() {
		s, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("error: %s", err.Error())
			os.Exit(1)
		}
		ctx.Append(openai.ChatMessageRoleUser, string(s))
	}

	if content != "" {
		ctx.Append(openai.ChatMessageRoleUser, content)
	}

	if content != "" || session.IsPipe() {
		_, err = Single(cfg, prof, ctx, isRestMode)
		if err != nil {
			fmt.Printf("error: %s", err.Error())
			os.Exit(1)
		}
	} else {
		StartDialog(cfg, prof, ctx, isRestMode, restore != "")
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
