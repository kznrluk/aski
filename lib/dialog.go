package lib

import (
	"bufio"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/kznrluk/aski/config"
	"github.com/sashabaranov/go-openai"
	"io"
	"os"
)

func StartDialog(cfg config.Config, profile config.Profile, ctx []openai.ChatCompletionMessage, isRestMode bool) {
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

		ctx = append(ctx, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Name:    profile.UserName,
			Content: input,
		})

		// fmt.Printf("%v", ctx)

		data := ""
		if isRestMode {
			d, err := restMode(oc, ctx, profile.Model)
			if err != nil {
				fmt.Printf(err.Error())
				continue
			}
			data = d
		} else {
			d, err := streamMode(oc, ctx, profile.Model)
			if err != nil {
				fmt.Printf(err.Error())
				continue
			}
			data = d
		}

		ctx = append(ctx, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: data,
		})

		fmt.Printf("\n\n%s@%s> ", profile.UserName, profile.ProfileName)
	}
}

func Single(cfg config.Config, profile config.Profile, ctx []openai.ChatCompletionMessage, isRestMode bool) (string, error) {
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

func restMode(oc *openai.Client, ctx []openai.ChatCompletionMessage, model string) (string, error) {
	yellow := color.New(color.FgHiYellow).SprintFunc()

	data := ""
	resp, err := oc.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: ctx,
		},
	)

	if err != nil {
		return "", err
	}
	fmt.Printf("%s", yellow(resp.Choices[0].Message.Content))
	data = resp.Choices[0].Message.Content

	return data, nil
}

func streamMode(oc *openai.Client, ctx []openai.ChatCompletionMessage, model string) (string, error) {
	yellow := color.New(color.FgHiYellow).SprintFunc()

	data := ""
	stream, err := oc.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: ctx,
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
