package lib

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/kznrluk/aski/chat"
	"github.com/kznrluk/aski/command"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/conv"
	"github.com/mattn/go-colorable"
	"github.com/sashabaranov/go-openai"
	"io"

	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/coloring"
	"github.com/nyaosorg/go-readline-ny/simplehistory"
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

	history := simplehistory.New()

	editor := &readline.Editor{
		PromptWriter: func(w io.Writer) (int, error) {
			return io.WriteString(w, "\u001B[0m"+profile.UserName+"@"+profile.ProfileName+"> ") // print `$ ` with cyan
		},
		Writer:         colorable.NewColorableStdout(),
		History:        history,
		Coloring:       &coloring.VimBatch{},
		HistoryCycling: true,
	}

	first := !restored
	for {
		input, err, interrupt := getInput(editor)
		history.Add(input)
		if interrupt || strings.HasPrefix(input, ":ex") {
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

		messages := cv.MessagesFromHead()
		if len(messages) > 0 {
			lastMessage := messages[len(messages)-1]
			showPendingHeader(openai.ChatMessageRoleAssistant, lastMessage)
		}

		data, err := chat.RetrieveResponse(isRestMode, cfg, cv)
		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		msg := cv.Append(openai.ChatMessageRoleAssistant, data)

		showMessageMeta(msg)
		if first {
			first = false

			if profile.Summarize {
				summary := chat.GetSummary(cfg, cv)
				cv.SetSummary(summary)
			}
		}
	}
}

func Single(cfg config.Config, ctx conv.Conversation, isRestMode bool) (string, error) {
	data, err := chat.RetrieveResponse(isRestMode, cfg, ctx)
	if err != nil {
		fmt.Printf(err.Error())
		return "", nil
	}

	return data, nil
}

func getInput(reader *readline.Editor) (string, error, bool) {
	sigintChan := make(chan os.Signal, 1)
	signal.Notify(sigintChan, os.Interrupt)

	inputChan := make(chan string, 1)
	go func() {
		input, err := reader.ReadLine(context.Background())
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
			msg := ctx.Last()
			fmt.Println(msg.Content)
		}

		return ctx, cont, nil
	}

	// direct send message
	ctx.Append(openai.ChatMessageRoleUser, input)

	return ctx, true, nil
}

func showPendingHeader(role string, to conv.Message) {
	yellow := color.New(color.FgHiYellow).SprintFunc()
	fmt.Print(yellow(fmt.Sprintf("\n------ [%s] -> %.*s", role, 6, to.Sha1)))
}

func showMessageMeta(msg conv.Message) {
	yellow := color.New(color.FgHiYellow).SprintFunc()
	fmt.Print(yellow(fmt.Sprintf("%.*s [%s] -> %.*s\n", 6, msg.Sha1, msg.Role, 6, msg.ParentSha1)))
}
