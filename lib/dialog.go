package lib

import (
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/kznrluk/aski/chat"
	"github.com/kznrluk/aski/command"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/conv"
	"github.com/mattn/go-colorable"
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

func StartDialog(cfg config.Config, cv conv.Conversation, isRestMode bool, restored bool) {
	if isRestMode {
		fmt.Printf("REST Mode \n")
	}

	history := simplehistory.New()

	profile := cv.GetProfile()
	editor := &readline.Editor{
		PromptWriter: func(w io.Writer) (int, error) {
			return io.WriteString(w, "\u001B[0m"+profile.UserName+"@"+profile.ProfileName+"> ") // print `$ ` with cyan
		},
		Writer:         colorable.NewColorableStdout(),
		History:        history,
		Coloring:       &coloring.VimBatch{},
		HistoryCycling: true,
	}

	editor.Init()
	fmt.Printf("Profile: %s, Model: %s \n", profile.ProfileName, profile.Model)

	cli := chat.ProvideChat(profile.Model, cfg)

	first := !restored
	for {
		fmt.Printf("\n")
		editor.PromptWriter = func(w io.Writer) (int, error) {
			return io.WriteString(w, fmt.Sprintf("%.*s > ", 6, cv.Last().Sha1))
		}

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

		last := cv.Last()
		yellow := color.New(color.FgHiYellow).SprintFunc()
		fmt.Print(yellow(fmt.Sprintf("\n%s -> [%.*s] \n", last.Role, 6, last.ParentSha1)))
		fmt.Print(fmt.Sprintf("%s", last.Content))
		fmt.Print(yellow(fmt.Sprintf(" [%.*s]\n", 6, last.Sha1)))

		messages := cv.MessagesFromHead()
		if len(messages) > 0 {
			lastMessage := messages[len(messages)-1]
			showPendingHeader(conv.ChatRoleAssistant, lastMessage)
		}

		fmt.Printf("\n")
		data, err := cli.Retrieve(cv, isRestMode)
		if err != nil {
			if errors.Is(err, chat.ErrCancelled) {
				_, _ = cv.ChangeHead(last.ParentSha1)
				continue
			}
			fmt.Printf("\n%s", err.Error())
			continue
		}

		msg := cv.Append(conv.ChatRoleAssistant, data)
		fmt.Print(yellow(fmt.Sprintf(" [%.*s]\n", 6, msg.Sha1)))
		if first {
			first = false
		}
	}
}

func OneShot(cfg config.Config, cv conv.Conversation, isRestMode bool) (string, error) {
	profile := cv.GetProfile()
	cli := chat.ProvideChat(profile.Model, cfg)

	data, err := cli.Retrieve(cv, isRestMode)

	fmt.Printf("\n") // in some cases, shell prompt delete the last line so we add a new line
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

	filename := fmt.Sprintf("%s.yaml", t.Format("20060102-150405"))

	homeDir, err := config.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
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

func appendMessage(input string, ctx conv.Conversation) (conv.Conversation, bool, error) {
	if len(input) > 0 && input[0] == ':' && input != ":exit" {
		ctx, cont, commandErr := command.Parse(input, ctx)
		if commandErr != nil {
			return ctx, false, commandErr
		}

		return ctx, cont, nil
	}

	// direct send message
	ctx.Append(conv.ChatRoleUser, input)

	return ctx, true, nil
}

func showPendingHeader(role string, to conv.Message) {
	yellow := color.New(color.FgHiYellow).SprintFunc()
	fmt.Print(yellow(fmt.Sprintf("\n%s -> [%.*s]", role, 6, to.Sha1)))
}
