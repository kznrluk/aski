package chat

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/conv"
	"github.com/sashabaranov/go-openai"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func RetrieveResponse(isRestMode bool, cfg config.Config, ctx conv.Conversation, model string) (string, error) {
	cancelCtx, cancelFunc := createCancellableContext()
	defer cancelFunc()

	oc := openai.NewClient(cfg.OpenAIAPIKey)
	if isRestMode {
		return Rest(cancelCtx, oc, ctx, model)
	}
	return Stream(cancelCtx, oc, ctx, model)
}

func GetSummary(cfg config.Config, conv conv.Conversation) string {
	oc := openai.NewClient(cfg.OpenAIAPIKey)

	c := ""
	for _, msg := range conv.GetMessages() {
		if msg.Role == openai.ChatMessageRoleSystem {
			continue
		}

		c += msg.Role + " says :" + msg.Content + "\n"
	}

	stream, err := oc.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: c,
					Name:    "User",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Give this conversation a short title in the language of the conversation in one line, without symbols.",
					Name:    "Aski",
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return ""
	}

	blue := color.New(color.FgHiBlue).SprintFunc()

	data := ""
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return ""
			}
		}

		fmt.Printf("%s", blue(resp.Choices[0].Delta.Content))
		data += resp.Choices[0].Delta.Content
	}

	fmt.Printf("\n")
	return data
}

func Rest(ctx context.Context, oc *openai.Client, conv conv.Conversation, model string) (string, error) {
	yellow := color.New(color.FgHiYellow).SprintFunc()

	resp, err := oc.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: conv.ToChatCompletionMessage(),
		},
	)

	if err != nil {
		return "", err
	}
	fmt.Printf("%s", yellow(resp.Choices[0].Message.Content))
	return resp.Choices[0].Message.Content, nil
}

func Stream(ctx context.Context, oc *openai.Client, conv conv.Conversation, model string) (string, error) {
	yellow := color.New(color.FgHiYellow).SprintFunc()

	stream, err := oc.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: conv.ToChatCompletionMessage(),
		},
	)

	if err != nil {
		return "", err
	}

	data := ""
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

	fmt.Printf("\n\n")
	return data, nil
}

func createCancellableContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT)

		select {
		case <-sigChan:
			cancel()
			println()
		case <-ctx.Done():
		}

		signal.Stop(sigChan)
	}()

	return ctx, cancel
}