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

func StartDialog(cfg config.Config, profile config.Profile, cv conv.Conversation, isRestMode bool, restored bool) {
	if isRestMode {
		fmt.Printf("REST Mode \n")
	}

	reader := bufio.NewReader(os.Stdin)
	first := !restored
	for {
		printPrompt(profile)

		input, err, interrupt := getInput(reader)
		if interrupt || input == ":exit" {
			if profile.AutoSave && !first {
				fmt.Printf("\nSaving conversation... ")
				fn, err := saveConversation(cv)
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

		cv, cont, commandErr := appendMessage(input, cv)
		if commandErr != nil {
			fmt.Printf("error: %v\n", commandErr)
		}

		if !cont {
			continue
		}

		data, err := chat.RetrieveResponse(isRestMode, cfg, cv, profile.Model)
		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		cv.Append(openai.ChatMessageRoleAssistant, data)
		if first {
			first = false

			if profile.Summarize {
				summary := chat.GetSummary(cfg, cv)
				cv.SetSummary(summary)
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
	if conv.GetSummary() != "" {
		escapedSummary += "_"
		escapedSummary += cleanFilenameElement(filepath.Clean(conv.GetSummary()))
	}
	filename := fmt.Sprintf("%s%s.yaml", t.Format("20060102-150405"), escapedSummary)

	configDir, err := config.GetHistoryDir()
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

func cleanFilenameElement(input string) string {
	invalidCharacters := [...]string{"/", "\\", "?", "%", "*", "|", "<", ">"}
	output := input
	for _, char := range invalidCharacters {
		output = strings.Replace(output, char, "-", -1)
	}
	return output
}

func appendMessage(input string, ctx conv.Conversation) (conv.Conversation, bool, error) {
	if len(input) > 0 && input[0] == ':' && input != ":exit" {
		ctx, cont, commandErr := command.Parse(input, ctx)
		if commandErr != nil {
			return ctx, false, commandErr
		}

		if cont {
			fmt.Print(ctx.Last().Content)
		}

		return ctx, cont, nil
	}

	// no command
	ctx.Append(openai.ChatMessageRoleUser, input)
	return ctx, true, nil
}

func printPrompt(profile config.Profile) {
	fmt.Printf("%s@%s> ", profile.UserName, profile.ProfileName)
}
