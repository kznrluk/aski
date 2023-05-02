package lib

import (
	"bufio"
	"fmt"
	"github.com/kznrluk/aski/chat"
	"github.com/kznrluk/aski/command"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/conv"
	"github.com/sashabaranov/go-openai"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

func StartDialog(cfg config.Config, profile config.Profile, conv conv.Conversation, isRestMode bool) {
	if isRestMode {
		fmt.Printf("REST Mode \n")
	}

	reader := bufio.NewReader(os.Stdin)
	first := true
	for {
		printPrompt(profile)

		input, err, interrupt := getInput(reader)
		if interrupt {
			if cfg.AutoSave && !first {
				fmt.Printf("\nSaving conversation... ")
				fn, err := saveConversation(conv)
				if err != nil {
					fmt.Printf("\n error saving conversation: %v\n", err)
					os.Exit(1)
				}
				fmt.Println(fn)
			}
			os.Exit(0)
		}

		if err != nil {
			fmt.Printf("Error Occured: %v\n", err)
			continue
		}

		if input == "" {
			continue
		}

		input, cont, commandErr := parseCommand(input, conv)
		if commandErr != nil {
			fmt.Printf("Command error: %v\n", commandErr)
		}

		if !cont {
			continue
		}

		conv.Append(openai.ChatMessageRoleUser, input)

		data, err := chat.RetrieveResponse(isRestMode, cfg, conv, profile.Model)
		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		conv.Append(openai.ChatMessageRoleAssistant, data)
		if first {
			first = false

			if cfg.Summarize {
				summary := chat.GetSummary(cfg, conv)
				conv.SetSummary(summary)
			}
		}
	}
}

func Single(cfg config.Config, profile config.Profile, ctx conv.Conversation, isRestMode bool) (string, error) {
	data, err := chat.RetrieveResponse(isRestMode, cfg, ctx, profile.Model)
	if err != nil {
		fmt.Printf(err.Error())
		return "", nil
	}

	return data, nil
}

func getInput(reader *bufio.Reader) (string, error, bool) {
	sigintChan := make(chan os.Signal, 1)
	signal.Notify(sigintChan, os.Interrupt)

	inputChan := make(chan string, 1)
	go func() {
		input, err := reader.ReadString('\n')
		if err != nil {
			inputChan <- ""
		} else {
			inputChan <- strings.TrimSpace(input)
		}
	}()

	select {
	case input := <-inputChan:
		return input, nil, false
	case <-sigintChan:
		return "", nil, true
	}
}

func saveConversation(conv conv.Conversation) (string, error) {
	t := time.Now()
	escapedSummary := ""
	if conv.Summary() != "" {
		escapedSummary += "_"
		escapedSummary += filepath.Clean(conv.Summary())
	}
	filename := fmt.Sprintf("%s%s.yaml", t.Format("20060102-150405"), escapedSummary)

	homeDir, err := config.GetHomeDir()
	if err != nil {
		return filename, err
	}

	configDir := filepath.Join(homeDir, ".aski", "history")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return filename, err
	}

	yamlString, err := conv.ToYAML()
	if err != nil {
		return filename, err
	}

	filePath := filepath.Join(configDir, filename)
	err = os.WriteFile(filePath, yamlString, 0600)
	if err != nil {
		return filename, err
	}

	return filename, nil
}

func parseCommand(input string, ctx conv.Conversation) (string, bool, error) {
	if len(input) > 0 && input[0] == ':' {
		str, cont, commandErr := command.Parse(input, ctx)
		if commandErr != nil {
			return "", false, commandErr
		}

		fmt.Println(str)
		return str, cont, nil
	}

	return input, true, nil
}

func printPrompt(profile config.Profile) {
	fmt.Printf("%s@%s> ", profile.UserName, profile.ProfileName)
}
