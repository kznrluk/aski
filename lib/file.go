package lib

import (
	"fmt"
	"github.com/kznrluk/aski/config"
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

func isBinary(contents []byte) bool {
	for _, ch := range contents {
		if ch == 0 {
			return true
		}
	}
	return false
}

func File(cmd *cobra.Command, args []string) {
	profileTarget, err := cmd.Flags().GetString("profile")
	isRestMode, _ := cmd.Flags().GetBool("rest")
	content, _ := cmd.Flags().GetString("content")
	system, _ := cmd.Flags().GetString("system")

	checkAPIKey()
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	var fileContents []FileContents
	for _, arg := range args {
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

	for _, f := range fileContents {
		ctx = append(ctx, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Name:    prof.UserName,
			Content: fmt.Sprintf("Path: `%s`\n ```\n%s```", f.Path, f.Contents),
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
