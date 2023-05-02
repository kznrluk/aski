package command

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/kznrluk/aski/ctx"
	"strings"
)

func Parse(input string, ctx ctx.Context) error {
	trimmedInput := strings.TrimSpace(input)
	commands := strings.Split(trimmedInput, " ")

	if commands[0] == ":context" {
		showContext(ctx)
		return nil
	} else if commands[0] == ":summary" {
		showSummary(ctx)
		return nil
	} else if commands[0] == ":move" {
		_ = changeHead(commands[1], ctx)
		return nil
	}

	return fmt.Errorf("unknown command")
}

func changeHead(sha1Partial string, context ctx.Context) error {
	if sha1Partial == "" {
		return fmt.Errorf("No SHA1 partial provided")
	}
	msg, err := context.ChangeHead(sha1Partial)
	if err != nil {
		return err
	}

	yellow := color.New(color.FgHiYellow).SprintFunc()
	blue := color.New(color.FgHiBlue).SprintFunc()
	fmt.Printf("%s %s\n", yellow(fmt.Sprintf("%.*s [%s]", 6, msg.Sha1, msg.Role)), blue("Head"))
	for _, context := range strings.Split(msg.Content, "\n") {
		fmt.Printf("  %s\n", context)
	}

	return nil
}

func showContext(ctx ctx.Context) {
	fmt.Println("Conversation context:")

	yellow := color.New(color.FgHiYellow).SprintFunc()
	blue := color.New(color.FgHiBlue).SprintFunc()
	for _, msg := range ctx.Messages() {
		Head := ""
		if msg.Head {
			Head = "Head"
		}
		fmt.Printf("%s %s\n", yellow(fmt.Sprintf("%.*s [%s]", 6, msg.Sha1, msg.Role)), blue(Head))

		for _, context := range strings.Split(msg.Content, "\n") {
			fmt.Printf("  %s\n", context)
		}

		fmt.Printf("\n")
	}
}

func showSummary(ctx ctx.Context) {
	summary := ctx.Summary()
	fmt.Printf(summary)
}
