package conv

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/session"
	"github.com/sashabaranov/go-openai"
	"strings"
)

type (
	Conversation interface {
		GetMessages() []Message
		MessagesFromHead() []Message
		SetSummary(summary string)
		GetSummary() string
		Append(role string, message string)
		ToChatCompletionMessage() []openai.ChatCompletionMessage
		ChangeHead(sha string) (Message, error)
		ToYAML() ([]byte, error)
	}

	conv struct {
		UserName string
		Model    string
		Summary  string
		Messages []Message
	}

	Message struct {
		Sha1       string
		ParentSha1 string
		Role       string
		Content    string `yaml:"content,literal"`
		UserName   string
		Head       bool
	}
)

func (c conv) GetMessages() []Message {
	return c.Messages
}

func (c *conv) SetSummary(summary string) {
	c.Summary = summary
}

func (c conv) GetSummary() string {
	return c.Summary
}

func (c *conv) Append(role string, message string) {
	parent := "ROOT"
	for i, m := range c.Messages {
		if m.Head {
			parent = m.Sha1
		}
		c.Messages[i].Head = false
	}

	sha := CalculateSHA1([]string{role, message, parent})

	msg := Message{
		Sha1:       sha,
		ParentSha1: parent,
		Role:       role,
		Content:    message,
		Head:       true,
	}

	if role == openai.ChatMessageRoleUser {
		msg.UserName = c.UserName
	}

	c.Messages = append(c.Messages, msg)
}

func (c *conv) ChangeHead(sha1Partial string) (Message, error) {
	foundSha := false
	foundMessageIndex := -1

	for i, message := range c.Messages {
		if strings.HasPrefix(message.Sha1, sha1Partial) {
			foundSha = true
			foundMessageIndex = i
			break
		}
	}

	if foundSha {
		for i := range c.Messages {
			c.Messages[i].Head = i == foundMessageIndex
		}
		return c.Messages[foundMessageIndex], nil
	}
	return Message{}, fmt.Errorf("no message found with provided sha1Partial: %s", sha1Partial)
}

func (c conv) MessagesFromHead() []Message {
	foundHead := false
	currentHead := ""
	for !foundHead {
		for _, message := range c.Messages {
			if message.Head {
				foundHead = true
				currentHead = message.Sha1
				break
			}
		}

		if !foundHead {
			break
		}

		messageChain := []Message{}
		for currentHead != "" {
			for i, message := range c.Messages {
				if message.Sha1 == currentHead {
					currentHead = message.ParentSha1
					messageChain = append(messageChain, message)
					break
				} else if i == len(c.Messages)-1 {
					currentHead = ""
				}
			}
		}

		for i, j := 0, len(messageChain)-1; i < j; i, j = i+1, j-1 {
			messageChain[i], messageChain[j] = messageChain[j], messageChain[i]
		}

		return messageChain
	}

	return []Message{}
}

func (c conv) ToChatCompletionMessage() []openai.ChatCompletionMessage {
	var chatMessages []openai.ChatCompletionMessage

	for _, message := range c.MessagesFromHead() {
		chatMessages = append(chatMessages, openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	if session.Verbose() {
		for _, message := range chatMessages {
			fmt.Printf("[%s]: %.32s\n", message.Role, message.Content)
		}
	}

	return chatMessages
}

func (c conv) ToYAML() ([]byte, error) {
	yamlBytes, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	return yamlBytes, nil
}

func NewConversation(profile config.Profile) Conversation {
	return &conv{
		UserName: profile.UserName,
		Model:    profile.Model,
		Summary:  "",
		Messages: []Message{},
	}
}

func FromYAML(yamlBytes []byte) (Conversation, error) {
	var c conv
	err := yaml.Unmarshal(yamlBytes, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func CalculateSHA1(stringsArray []string) string {
	combinedString := strings.Join(stringsArray, "")
	hasher := sha1.New()
	hasher.Write([]byte(combinedString))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}
