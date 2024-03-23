package conv

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/kznrluk/aski/config"
	"github.com/kznrluk/aski/session"
	"github.com/kznrluk/aski/util"
	"github.com/kznrluk/go-anthropic"
	"github.com/sashabaranov/go-openai"
	"strings"
)

type (
	Conversation interface {
		GetMessages() []Message
		GetMessageFromSha1(sha1partial string) (Message, error)
		Last() Message
		MessagesFromHead() []Message
		Append(role string, message string) Message
		SetSystem(message string)
		GetSystem() string
		SetProfile(profile config.Profile) error
		Modify(m Message) error
		ToOpenAIMessage() []openai.ChatCompletionMessage
		ToAnthropicMessage() []anthropic.Message
		ChangeHead(sha string) (Message, error)
		GetProfile() config.Profile
		ToYAML() ([]byte, error)
	}

	conv struct {
		Profile  config.Profile
		System   string
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

const (
	ChatRoleUser      = "user"
	ChatRoleAssistant = "assistant"
)

func (c conv) GetMessages() []Message {
	return c.Messages
}

func (c conv) Last() Message {
	if len(c.Messages) == 0 {
		return c.convertSystemToMessage()
	}
	return c.Messages[len(c.Messages)-1]
}

func (c *conv) SetSystem(text string) {
	c.System = text
}

func (c conv) GetSystem() string {
	return c.System
}

func (c *conv) Modify(m Message) error {
	for i, message := range c.Messages {
		if message.Sha1 == m.Sha1 {
			c.Messages[i] = m
			return nil
		}
	}

	return fmt.Errorf("no message found with provided sha1: %s", m.Sha1)
}

func (c *conv) Append(role string, message string) Message {
	parent := "ROOT"
	for i, m := range c.Messages {
		if m.Head {
			parent = m.Sha1
		}
		c.Messages[i].Head = false
	}

	sha := CalculateSHA1([]string{role, message, parent})

	if c.Profile.DiceRoll != "" {
		result, err := util.RollDice(c.Profile.DiceRoll)
		if err != nil {
			panic(err) // profile validation should have caught this
		}
		message = fmt.Sprintf("%s\n DiceRoll %s: %d", message, c.Profile.DiceRoll, result)
	}

	msg := Message{
		Sha1:       sha,
		ParentSha1: parent,
		Role:       role,
		Content:    message,
		Head:       true,
	}

	if role == ChatRoleUser {
		msg.UserName = c.Profile.UserName
	}

	c.Messages = append(c.Messages, msg)

	return msg
}

func (c *conv) GetMessageFromSha1(sha1partial string) (Message, error) {
	for _, message := range c.Messages {
		if strings.HasPrefix(message.Sha1, sha1partial) {
			return message, nil
		}
	}
	return Message{}, fmt.Errorf("no message found with provided sha1partial: %s", sha1partial)
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

func (c conv) ToOpenAIMessage() []openai.ChatCompletionMessage {
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

func (c conv) ToAnthropicMessage() []anthropic.Message {
	var chatMessages []anthropic.Message

	// NOTE: Anthropic does not include system messages in the conversation
	for _, message := range c.MessagesFromHead() {
		var role string

		if message.Role == ChatRoleUser {
			role = anthropic.ChatMessageRoleUser
		} else if message.Role == ChatRoleAssistant {
			role = anthropic.ChatMessageRoleAssistant
		} else {
			panic(fmt.Sprintf("unknown role: %s", message.Role))
		}
		chatMessages = append(chatMessages, anthropic.Message{
			Role:    role,
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

func (c conv) GetProfile() config.Profile {
	return c.Profile
}

func (c *conv) SetProfile(profile config.Profile) error {
	c.Profile = profile
	return nil
}

func (c conv) convertSystemToMessage() Message {
	return Message{
		Sha1:       CalculateSHA1([]string{c.System}),
		ParentSha1: "ROOT",
		Role:       "system",
		Content:    c.System,
		Head:       false,
	}
}

func NewConversation(profile config.Profile) Conversation {
	return &conv{
		Profile:  profile,
		Messages: []Message{},
	}
}

func FromYAML(yamlBytes []byte) (Conversation, error) {
	var c conv
	err := yaml.Unmarshal(yamlBytes, &c)
	if err != nil {
		return nil, err
	}

	// Decode tab escape sequences
	for i, message := range c.Messages {
		c.Messages[i].Content = strings.ReplaceAll(message.Content, "\\t", "\t")
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
