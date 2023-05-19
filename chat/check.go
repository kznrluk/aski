package chat

import (
	"bufio"
	"context"
	"fmt"
	"github.com/kznrluk/aski/config"
	"github.com/sashabaranov/go-openai"
	"os"
	"strings"
)

// PromptGetAPIKey prompts the user to enter an API key and saves it to the configuration file.
func PromptGetAPIKey(cfg config.Config) {
	fmt.Printf("Please generate an API key from this URL. Currently, the configuration file is saved in plaintext. \nhttps://platform.openai.com/account/api-keys\n")
	fmt.Printf("\t OpenAI API Key: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		textRaw := scanner.Text()
		text := strings.TrimSpace(textRaw)

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
