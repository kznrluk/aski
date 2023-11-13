package lib

import (
	"fmt"
	"github.com/kznrluk/aski/chat"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/conv"
	"github.com/kznrluk/aski/file"
	"github.com/kznrluk/aski/session"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

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

	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	if cfg.OpenAIAPIKey == "" {
		chat.PromptGetAPIKey(&cfg)
	}

	prof, err := config.GetProfile(cfg, profileTarget)
	if err != nil {
		fmt.Printf("error getting profile: %v\n. using default profile.", err)
		prof = config.InitialProfile()
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
		ctx.Append(openai.ChatMessageRoleSystem, prof.SystemContext)

		if len(fileGlobs) != 0 {
			fileContents := file.GetFileContents(fileGlobs)
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

	if content != "" {
		_, err = Single(cfg, ctx, isRestMode)
		if err != nil {
			fmt.Printf("error: %s", err.Error())
			os.Exit(1)
		}
	} else {
		StartDialog(cfg, ctx, isRestMode, restore != "")
	}
}
