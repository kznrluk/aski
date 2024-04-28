package chat

import (
	"context"
	"errors"
	"fmt"
	"github.com/kznrluk/aski/conv"
	"github.com/kznrluk/go-anthropic"
	"io"
)

type (
	ap struct {
		ac *anthropic.Client
	}
)

func (a ap) Retrieve(conv conv.Conversation, useRest bool) (string, error) {
	if useRest {
		return a.RetrieveRest(conv)
	}
	return a.RetrieveStream(conv)
}

func (a ap) RetrieveRest(conv conv.Conversation) (string, error) {
	cancelCtx, cancelFunc := createCancellableContext()
	defer cancelFunc()
	return a.rest(cancelCtx, conv)
}

func (a ap) RetrieveStream(conv conv.Conversation) (string, error) {
	cancelCtx, cancelFunc := createCancellableContext()
	defer cancelFunc()
	return a.stream(cancelCtx, conv)
}

func (a ap) rest(ctx context.Context, conv conv.Conversation) (string, error) {
	messages := conv.ToAnthropicMessage()
	model := conv.GetProfile().Model
	rest, err := a.ac.CreateMessage(
		ctx,
		anthropic.MessageRequest{
			MaxTokens: 4096,
			Model:     model,
			System:    conv.GetSystem(),
			Messages:  messages,
		},
	)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			return "", ErrCancelled
		}
		return "", err
	}
	if len(rest.Content) == 0 {
		return "", fmt.Errorf("no content")
	}
	fmt.Printf("%s", rest.Content[0].Text)
	return rest.Content[0].Text, nil
}

func (a ap) stream(ctx context.Context, conv conv.Conversation) (string, error) {
	messages := conv.ToAnthropicMessage()
	model := conv.GetProfile().Model
	stream, err := a.ac.CreateMessageStream(
		ctx,
		anthropic.MessageRequest{
			MaxTokens: 4096,
			Model:     model,
			System:    conv.GetSystem(),
			Messages:  messages,
		},
	)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			return "", ErrCancelled
		}
		return "", err
	}

	data := ""
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else if errors.Is(err, context.Canceled) {
				return "", ErrCancelled
			} else {
				fmt.Printf("%s", err.Error())
				return "", err
			}
		}

		fmt.Printf("%s", resp.Delta.Text)
		data += resp.Delta.Text
	}
	return data, nil
}

func NewAnthropic(key string) Chat {
	return ap{ac: anthropic.NewClient(key)}
}
