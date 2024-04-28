package chat

import (
	"context"
	"errors"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/conv"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type (
	Chat interface {
		Retrieve(conv conv.Conversation, useRest bool) (string, error)
		RetrieveRest(conv conv.Conversation) (string, error)
		RetrieveStream(conv conv.Conversation) (string, error)
	}
)

var (
	ErrCancelled = errors.New("cancelled")
)

func ProvideChat(model string, cfg config.Config) Chat {
	if strings.HasPrefix(model, "claude") {
		return NewAnthropic(cfg.AnthropicAPIKey)
	}

	return NewOpenAI(cfg.OpenAIAPIKey)
}

func createCancellableContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT)

		select {
		case <-sigChan:
			println()
			cancel()
		case <-ctx.Done():
		}

		signal.Stop(sigChan)
	}()

	return ctx, cancel
}
