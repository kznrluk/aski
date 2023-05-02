package ctx

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"github.com/goccy/go-yaml"
	"github.com/kznrluk/aski/config"
	"github.com/sashabaranov/go-openai"
	"strings"
)

type (
	Context interface {
		Messages() []Message
		Summary() string
		Append(role string, message string)
		ToChatCompletionMessage() []openai.ChatCompletionMessage
		ChangeHead(sha string) (Message, error)
		ToYAML() ([]byte, error)
	}

	context struct {
		profile  config.Profile
		config   config.Config
		model    string    `yaml:"Model"`
		summary  string    `yaml:"Summary"`
		messages []Message `yaml:"Messages"`
	}

	Message struct {
		Sha1       string
		ParentSha1 string
		Role       string
		Content    string
		UserName   string
		Head       bool
	}
)

func (c context) Messages() []Message {
	return c.messages
}

func (c context) Summary() string {
	userMessageCount := 0
	for _, message := range c.messages {
		if message.Role == "User" {
			userMessageCount++
		}
	}

	if userMessageCount > 0 {
		// todo
	}

	return c.summary
}

func (c *context) Append(role string, message string) {
	parent := ""
	for i, m := range c.messages {
		if m.Head {
			parent = m.Sha1
		}
		c.messages[i].Head = false
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
		msg.UserName = c.profile.UserName
	}

	c.messages = append(c.messages, msg)
}

func (c *context) ChangeHead(sha1Partial string) (Message, error) {
	found := false
	foundIndex := -1

	for i, message := range c.messages {
		if strings.HasPrefix(message.Sha1, sha1Partial) {
			if found {
				foundIndex = i
			} else {
				found = true
				foundIndex = i
			}
		}

		// Headフラグを切り替える
		c.messages[i].Head = (i == foundIndex)
	}

	if !found {
		return Message{}, errors.New("No message found")
	}

	return c.messages[foundIndex], nil
}

func (c context) ToChatCompletionMessage() []openai.ChatCompletionMessage {
	chatMessages := make([]openai.ChatCompletionMessage, len(c.messages))

	for i, message := range c.messages {
		chatMessages[i] = openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		}
	}

	return chatMessages
}

func (c context) ToYAML() ([]byte, error) {
	yamlBytes, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	return yamlBytes, nil
}

func NewContext(profile config.Profile, config config.Config) Context {
	return &context{
		profile:  profile,
		config:   config,
		model:    profile.Model,
		summary:  "",
		messages: []Message{},
	}
}

func CalculateSHA1(stringsArray []string) string {
	combinedString := strings.Join(stringsArray, "")
	hasher := sha1.New()
	hasher.Write([]byte(combinedString))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}
