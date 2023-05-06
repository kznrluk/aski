package command

import (
	"fmt"
	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/conv"
	"github.com/sashabaranov/go-openai"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type cmd struct {
	name        string
	description string
}

var availableCommands = []cmd{
	{
		name:        ":history",
		description: "Show conversation history.",
	},
	{
		name:        ":summary",
		description: "Show conversation summary.",
	},
	{
		name:        ":move",
		description: "Change HEAD to another message.",
	},
	{
		name:        ":config",
		description: "Open configuration directory.",
	},
	{
		name: ":editor",
		description: "Open an external text editor to add new message.\n" +
			"  :editor sha1   - Edit the argument message and continue the conversation.\n" +
			"  :editor latest - Edits the nearest own statement from HEAD.",
	},
	{
		name: ":modify sha1",
		description: "Modify the past conversation. HEAD does not move.\n" +
			"                   Past conversations will be modified from the next transmission.",
	},
	{
		name:        ":exit",
		description: "Exit the program.",
	},
}

func matchCommand(input string) (string, bool) {
	matched := false
	var matchedCmd string

	for _, cmd := range availableCommands {
		if strings.HasPrefix(cmd.name, input) {
			if matched {
				return "", false // ambiguous input, more than one command matches
			}
			matched = true
			matchedCmd = cmd.name
		}
	}

	return matchedCmd, matched
}

func unknownCommand() string {
	output := "unknown command.\n\n"
	for _, cmd := range availableCommands {
		output += fmt.Sprintf("  %-14s - %s\n", cmd.name, cmd.description)
	}

	return output
}

func Parse(input string, conv conv.Conversation) (conv.Conversation, bool, error) {
	trimmedInput := strings.TrimSpace(input)
	commands := strings.Split(trimmedInput, " ")

	matchedCmd, found := matchCommand(commands[0])
	if !found {
		return nil, false, fmt.Errorf(unknownCommand())
	}
	commands[0] = matchedCmd

	if commands[0] == ":history" {
		showContext(conv)
		return nil, false, nil
	} else if commands[0] == ":summary" {
		showSummary(conv)
		return nil, false, nil
	} else if commands[0] == ":move" {
		err := changeHead(commands[1], conv)
		return nil, false, err
	} else if commands[0] == ":config" {
		_ = config.OpenConfigDir()
		return nil, false, nil
	} else if commands[0] == ":editor" {
		trim := ""
		if len(commands) > 1 {
			trim = strings.TrimSpace(commands[1])
		}

		if trim == "" {
			return newMessage(conv)
		}

		return editMessage(conv, trim)
	} else if commands[0] == ":modify sha1" {
		return modifyMessage(conv, commands[1])
	}

	return nil, false, fmt.Errorf("unknown command")
}

func changeHead(sha1Partial string, context conv.Conversation) error {
	if sha1Partial == "" {
		return fmt.Errorf("No SHA1 partial provided")
	}
	msg, err := context.ChangeHead(sha1Partial)
	if err != nil {
		return err
	}

	yellow := color.New(color.FgHiYellow).SprintFunc()
	blue := color.New(color.FgHiBlue).SprintFunc()
	fmt.Printf("%s %s\n", yellow(yellow(fmt.Sprintf("%.*s [%s] -> %.*s", 6, msg.Sha1, msg.Role, 6, msg.ParentSha1))), blue("Head"))
	for _, context := range strings.Split(msg.Content, "\n") {
		fmt.Printf("  %s\n", context)
	}

	return nil
}

func showContext(conv conv.Conversation) {
	yellow := color.New(color.FgHiYellow).SprintFunc()
	blue := color.New(color.FgHiBlue).SprintFunc()

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)

	for _, msg := range conv.GetMessages() {
		head := ""
		if msg.Head {
			head = "Head"
		}
		fmt.Printf("%s %s\n", yellow(fmt.Sprintf("%.*s [%s] -> %.*s", 6, msg.Sha1, msg.Role, 6, msg.ParentSha1)), blue(head))

		out, err := r.Render(msg.Content)
		if err != nil {
			fmt.Printf("error: create markdown failed: %s", err.Error())
		}

		out = strings.TrimSpace(out)
		for _, context := range strings.Split(out, "\n") {
			fmt.Printf("%s\n", context)
		}

		fmt.Printf("\n")
	}
}

func showSummary(conv conv.Conversation) {
	blue := color.New(color.FgHiBlue).SprintFunc()
	fmt.Printf(blue(conv.GetSummary()))
}

func newMessage(conv conv.Conversation) (conv.Conversation, bool, error) {
	comments := "\n\n# Save and close editor to continue\n"
	s := conv.MessagesFromHead()
	for i := len(s) - 1; i >= 0; i-- {
		msg := s[i]
		head := ""
		if msg.Head {
			head = "Head"
		}

		d := fmt.Sprintf("#\n# %.*s -> %.*s [%s] %s\n", 6, msg.Sha1, 6, msg.ParentSha1, msg.Role, head)
		for _, context := range strings.Split(msg.Content, "\n") {
			d += fmt.Sprintf("#   %s\n", context)
		}
		comments += d
	}

	result, err := openEditor(comments)
	if err != nil {
		return nil, false, fmt.Errorf("failed to open editor: %v", err)
	}

	if result == "" {
		return conv, false, nil
	}

	conv.Append(openai.ChatMessageRoleUser, result)
	return conv, true, nil
}

func modifyMessage(cv conv.Conversation, sha1 string) (conv.Conversation, bool, error) {
	trimmedSha1 := strings.TrimSpace(sha1)
	if trimmedSha1 == "" {
		return nil, false, fmt.Errorf("no SHA1 provided")
	}

	msg, err := cv.GetMessageFromSha1(trimmedSha1)
	if err != nil {
		return nil, false, fmt.Errorf("failed to edit message from SHA1: %v", err)
	}

	s := cv.MessagesFromHead()
	comments := msg.Content + "\n\n# Save and close editor to continue\n"
	for i := len(s) - 1; i >= 0; i-- {
		m := s[i]
		head := ""
		if m.Head {
			head = "Head"
		}

		d := fmt.Sprintf("#\n# %.*s -> %.*s [%s] %s\n", 6, m.Sha1, 6, m.ParentSha1, m.Role, head)
		for _, context := range strings.Split(m.Content, "\n") {
			d += fmt.Sprintf("#   %s\n", context)
		}
		comments += d
	}

	result, err := openEditor(comments)
	if err != nil {
		return nil, false, fmt.Errorf("failed to open editor: %v", err)
	}

	if result == "" {
		return cv, false, nil
	}

	if strings.TrimSpace(result) == strings.TrimSpace(msg.Content) {
		return cv, false, nil
	}

	msg.Content = result

	err = cv.Modify(msg)
	if err != nil {
		return nil, false, fmt.Errorf("failed to modify message: %v", err)
	}

	fmt.Printf("[%.6s] Modified. \n", msg.Sha1)
	return cv, false, nil
}

func editMessage(cv conv.Conversation, sha1 string) (conv.Conversation, bool, error) {
	trimmedSha1 := strings.TrimSpace(sha1)
	if trimmedSha1 == "" {
		return nil, false, fmt.Errorf("no SHA1 provided")
	}

	var msg conv.Message
	if strings.ToLower(trimmedSha1) == "latest" {
		msgs := cv.MessagesFromHead()
		for i := len(msgs) - 1; i >= 0; i-- {
			if msgs[i].Role == openai.ChatMessageRoleUser {
				msg = msgs[i]
				break
			}
		}

		if msg.Sha1 == "" {
			return nil, false, fmt.Errorf("no latest user message found")
		}
	} else {
		m, err := cv.GetMessageFromSha1(trimmedSha1)
		if err != nil {
			return nil, false, fmt.Errorf("failed to edit message from SHA1: %v", err)
		}

		if m.Role != openai.ChatMessageRoleUser {
			return nil, false, fmt.Errorf("cannot edit non-user message")
		}

		msg = m
	}

	s := cv.MessagesFromHead()
	comments := msg.Content + "\n\n# Save and close editor to continue\n"
	for i := len(s) - 1; i >= 0; i-- {
		m := s[i]
		head := ""
		if m.Head {
			head = "Head"
		}

		d := fmt.Sprintf("#\n# %.*s -> %.*s [%s] %s\n", 6, m.Sha1, 6, m.ParentSha1, m.Role, head)
		for _, context := range strings.Split(m.Content, "\n") {
			d += fmt.Sprintf("#   %s\n", context)
		}
		comments += d
	}

	result, err := openEditor(comments)
	if err != nil {
		return nil, false, fmt.Errorf("failed to open editor: %v", err)
	}

	if result == "" {
		return cv, false, nil
	}

	if strings.TrimSpace(result) == strings.TrimSpace(msg.Content) {
		return cv, false, nil
	}

	_, err = cv.ChangeHead(msg.ParentSha1)
	if err != nil {
		return nil, false, fmt.Errorf("failed to change head: %v", err)
	}

	cv.Append(msg.Role, result)
	return cv, true, nil
}

func openEditor(content string) (string, error) {
	tempDir, err := config.GetAskiDir()
	if err != nil {
		return "", fmt.Errorf("failed to get aski directory: %v", err)
	}
	tmpFile, err := os.CreateTemp(tempDir, "aski-editor-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create a temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		return "", fmt.Errorf("failed to write to the temp file: %v", err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad.exe"
		} else {
			editor = "vim"
		}
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// for vscode :)
	if strings.Contains(editor, "code") {
		cmd = exec.Command(editor, "--wait", tmpFile.Name())
	}

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to open editor: %v", err)
	}

	tmpFile.Close()

	rawContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read the edited content: %v", err)
	}

	result := ""
	for _, d := range strings.Split(string(rawContent), "\n") {
		if !strings.HasPrefix(d, "#") {
			result += d + "\n"
		}
	}
	result = strings.TrimSpace(result)
	if len(strings.TrimSpace(result)) == 0 {
		return "", nil
	}

	return result, nil
}
