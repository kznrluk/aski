package lib

import (
	"fmt"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/conv"
	"github.com/kznrluk/aski/file"
	"github.com/kznrluk/aski/session"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

func Aski(cmd *cobra.Command, args []string) {
	profileTarget, err := cmd.Flags().GetString("profile")
	isRestMode, _ := cmd.Flags().GetBool("rest")
	content, _ := cmd.Flags().GetString("content")
	model, _ := cmd.Flags().GetString("model")
	fileGlobs, _ := cmd.Flags().GetStringSlice("file")
	restore, _ := cmd.Flags().GetString("restore")
	verbose, _ := cmd.Flags().GetBool("verbose")
	session.SetVerbose(verbose)

	fileInfo, _ := os.Stdin.Stat()
	if (fileInfo.Mode() & os.ModeNamedPipe) != 0 {
		session.SetIsPipe(true)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	if cfg.OpenAIAPIKey == "" && cfg.AnthropicAPIKey == "" {
		configPath := config.MustGetAskiDir()
		fmt.Printf("No API key found. Please set your API key in %s/config.yaml\n", configPath)
		os.Exit(1)
	}

	prof, err := config.GetProfile(cfg, profileTarget)
	if err != nil {
		fmt.Printf("error getting profile: %v\n. using default profile.", err)
		prof = config.InitialProfile()
	}

	if model != "" {
		prof.Model = model
	}

	if strings.HasPrefix(prof.Model, "gpt") && cfg.OpenAIAPIKey == "" {
		fmt.Printf("OpenAIAPIKey is required for GPT model. Please set your API key in %s/config.yaml\n", config.MustGetAskiDir())
		os.Exit(1)
	}

	if strings.HasPrefix(prof.Model, "claude") && cfg.AnthropicAPIKey == "" {
		fmt.Printf("AnthropicAPIKey is required for Claude model. Please set your API key in %s/config.yaml\n", config.MustGetAskiDir())
		os.Exit(1)
	}

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

		if profileTarget != "" {
			fmt.Printf("WARN: Profile is ignored when loading restore.\n")
		}

		println("Restore conversations from " + fileName)
	} else {
		ctx = conv.NewConversation(prof)
		ctx.SetSystem(prof.SystemContext)

		if len(fileGlobs) != 0 {
			fileContents := file.GetFileContents(fileGlobs)
			for _, f := range fileContents {
				if content == "" && !session.IsPipe() {
					fmt.Printf("Append File: %s\n", f.Name)
				}
				ctx.Append(conv.ChatRoleUser, fmt.Sprintf("Path: `%s`\n ```\n%s```", f.Path, f.Contents))
			}
		}

		for _, i := range prof.Messages {
			switch strings.ToLower(i.Role) {
			case conv.ChatRoleUser:
				ctx.Append(conv.ChatRoleUser, i.Content)
			case conv.ChatRoleAssistant:
				ctx.Append(conv.ChatRoleAssistant, i.Content)
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
		ctx.Append(conv.ChatRoleUser, string(s))
	}

	if content != "" {
		ctx.Append(conv.ChatRoleUser, content)
	}

	if content != "" {
		_, err = OneShot(cfg, ctx, isRestMode)
		if err != nil {
			fmt.Printf("error: %s", err.Error())
			os.Exit(1)
		}
	} else {
		StartDialog(cfg, ctx, isRestMode, restore != "")
	}
}
