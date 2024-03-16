package chat

import (
	"context"
	"fmt"
	"github.com/kznrluk/aski/conv"
	"github.com/sashabaranov/go-openai"
	"io"
)

type (
	oai struct {
		oc *openai.Client
	}
)

func (o oai) Retrieve(conv conv.Conversation, useRest bool) (string, error) {
	if useRest {
		return o.RetrieveRest(conv)
	}
	return o.RetrieveStream(conv)
}

func (o oai) RetrieveRest(conv conv.Conversation) (string, error) {
	cancelCtx, cancelFunc := createCancellableContext()
	defer cancelFunc()
	return o.rest(cancelCtx, conv)
}

func (o oai) RetrieveStream(conv conv.Conversation) (string, error) {
	cancelCtx, cancelFunc := createCancellableContext()
	defer cancelFunc()
	return o.stream(cancelCtx, conv)
}

func (o oai) rest(ctx context.Context, conv conv.Conversation) (string, error) {
	profile := conv.GetProfile()
	customParams := profile.CustomParameters
	resp, err := o.oc.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:            profile.Model,
			Messages:         conv.ToOpenAIMessage(),
			ResponseFormat:   profile.GetResponseFormat(),
			MaxTokens:        customParams.MaxTokens,
			Temperature:      customParams.Temperature,
			TopP:             customParams.TopP,
			Stop:             customParams.Stop,
			PresencePenalty:  customParams.PresencePenalty,
			FrequencyPenalty: customParams.FrequencyPenalty,
			LogitBias:        customParams.LogitBias,
		},
	)

	if err != nil {
		return "", err
	}
	fmt.Printf("%s", resp.Choices[0].Message.Content)
	return resp.Choices[0].Message.Content, nil
}

func (o oai) stream(ctx context.Context, conv conv.Conversation) (string, error) {
	profile := conv.GetProfile()
	customParams := profile.CustomParameters
	stream, err := o.oc.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:            profile.Model,
			Messages:         conv.ToOpenAIMessage(),
			ResponseFormat:   profile.GetResponseFormat(),
			MaxTokens:        customParams.MaxTokens,
			Temperature:      customParams.Temperature,
			TopP:             customParams.TopP,
			Stop:             customParams.Stop,
			PresencePenalty:  customParams.PresencePenalty,
			FrequencyPenalty: customParams.FrequencyPenalty,
			LogitBias:        customParams.LogitBias,
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

		fmt.Printf("%s", resp.Choices[0].Delta.Content)
		data += resp.Choices[0].Delta.Content
	}
	return data, nil
}

func NewOpenAI(key string) Chat {
	return oai{oc: openai.NewClient(key)}
}
