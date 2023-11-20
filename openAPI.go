package main

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
)

const (
	MaxTokens = 100
)

type Key string

var key Key = ""

func createKey(key string) error {
	return os.WriteFile("key", []byte(key), 0644)
}

// getKey NOTE : Lazy load
func getKey() (Key, error) {
	// Key already loaded
	if key != "" {
		return key, nil
	}

	// Key file doesn't exist
	if _, err := os.Stat("key"); os.IsNotExist(err) {
		return "", errors.New("key not found")
	}

	// Key file exist but is invalid
	data, err := os.ReadFile("key")
	if err != nil {
		return "", err
	}
	tmpKey := strings.TrimRight(string(data), "\n")
	regex := regexp.MustCompile(`^sk-[a-zA-Z0-9]{48}$`)
	if !regex.MatchString(tmpKey) {
		return "", errors.New("key is invalid")
	}

	// Key file exist and is valid
	key = Key(tmpKey)
	return key, nil
}

// invalid TODO : invalid key on every error or need change
// invalid NOTE : Should never be used outside a function with a malloc
func (key *Key) invalid() bool {
	if *key == "" {
		return false
	}
	*key = ""
	return true
}

type OpenClient struct {
	client *openai.Client
}

var openClient = OpenClient{}

func GetClient() (*openai.Client, error) {
	if openClient.client == nil {
		key, err := getKey()
		if err != nil {
			return nil, err
		}
		openClient.client = openai.NewClient(string(key))
	}
	return openClient.client, nil
}

func (openClient *OpenClient) invalid() bool {
	ok := true
	if openClient.client == nil {
		ok = false
	}
	openClient.client = nil
	return ok
}

func (conv *Conversation) chatCompletionSize(question string, maxTokens int, model string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	newQuestion := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	}

	req := openai.ChatCompletionRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages:  append(conv.openaiMessages(), newQuestion),
		Stream:    false,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return err
	}

	// TODO : In which case it should not store the question and the answer ?
	conv.Messages = append(conv.Messages, userOpenaiMessage(newQuestion).toMessage())
	conv.Messages = append(conv.Messages, gptMessage(resp.Choices[0]).toMessage(model))
	conv.HasChange = true
	conv.LastModel = model

	// NOTE : is there a choices 0 in case of err ?
	return err
}

func (conv *Conversation) chatCompletion(question string, model string) error {
	return conv.chatCompletionSize(question, MaxTokens, model)
}

func (conv *Conversation) chatCompletionNoModel(question string) error {
	return conv.chatCompletion(question, conv.LastModel)
}
