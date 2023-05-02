package lib

import (
	"bufio"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/kznrluk/aski/command"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/ctx"
	"github.com/sashabaranov/go-openai"
	"io"
	"os"
)

func StartDialog(cfg config.Config, profile config.Profile, ctx ctx.Context, isRestMode bool) {
	oc := openai.NewClient(cfg.OpenAIAPIKey)

	if isRestMode {
		fmt.Printf("REST Mode \n")
	}

	fmt.Printf("%s@%s> ", profile.UserName, profile.ProfileName)
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error Occured: %v\n", err)
			continue
		}

		if len(input) == 0 {
			printPrompt(profile)
			continue
		}

		if len(input) > 0 && input[0] == ':' {
			commandErr := command.Parse(input, ctx)
			if commandErr != nil {
				fmt.Printf("Command error: %v\n", commandErr)
			}

			printPrompt(profile)
			continue
		}

		ctx.Append(openai.ChatMessageRoleSystem, input)

		data := ""
		if isRestMode {
			data, err = restMode(oc, ctx, profile.Model)
			if err != nil {
				fmt.Printf(err.Error())
				continue
			}
		} else {
			data, err = streamMode(oc, ctx, profile.Model)
			if err != nil {
				fmt.Printf(err.Error())
				continue
			}
		}

		ctx.Append(openai.ChatMessageRoleAssistant, data)
		printPrompt(profile)
	}
}

func Single(cfg config.Config, profile config.Profile, ctx ctx.Context, isRestMode bool) (string, error) {
	oc := openai.NewClient(cfg.OpenAIAPIKey)

	data := ""
	if isRestMode {
		d, err := restMode(oc, ctx, profile.Model)
		if err != nil {
			fmt.Printf(err.Error())
			return "", nil
		}
		data = d
	} else {
		d, err := streamMode(oc, ctx, profile.Model)
		if err != nil {
			fmt.Printf(err.Error())
			return "", nil
		}
		data = d
	}

	return data, nil
}

func restMode(oc *openai.Client, ctx ctx.Context, model string) (string, error) {
	yellow := color.New(color.FgHiYellow).SprintFunc()

	data := ""
	resp, err := oc.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: ctx.ToChatCompletionMessage(),
		},
	)

	if err != nil {
		return "", err
	}
	fmt.Printf("%s, %s", yellow(resp.Choices[0].Message.Content))
	data = resp.Choices[0].Message.Content

	return data, nil
}

func streamMode(oc *openai.Client, ctx ctx.Context, model string) (string, error) {
	yellow := color.New(color.FgHiYellow).SprintFunc()

	data := ""
	stream, err := oc.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: ctx.ToChatCompletionMessage(),
		},
	)

	if err != nil {
		return "", err
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "", err
			}
		}

		fmt.Printf("%s", yellow(resp.Choices[0].Delta.Content))
		data += resp.Choices[0].Delta.Content
	}

	return data, nil
}

func printPrompt(profile config.Profile) {
	fmt.Printf("\n\n%s@%s> ", profile.UserName, profile.ProfileName)
}
