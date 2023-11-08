package main

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"os"
)

const (
	MaxTokens = 100
)

type (
	Conversation struct {
		ID        string                         `json:"id"`
		Model     string                         `json:"model"`
		Name      string                         `json:"name"`
		HasChange bool                           `json:"has_change"`
		Messages  []openai.ChatCompletionMessage `json:"messages"`
	}

	Key string
)

var key Key = ""

func GetKey() (Key, error) {
	if key == "" {
		dat, err := os.ReadFile("key")
		if err != nil {
			return "", err
		}
		key = Key(dat)
	}
	return key, nil
}

// Invalid TODO : Invalid key on every error or need change
// Should never be used outside a function with a malloc
func (key *Key) Invalid() bool {
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
		key, err := GetKey()
		if err != nil {
			return nil, err
		}
		openClient.client = openai.NewClient(string(key))
	}
	return openClient.client, nil
}

func (openClient *OpenClient) Invalid() bool {
	ok := true
	if openClient.client == nil {
		ok = false
	}
	openClient.client = nil
	return ok
}

func (conv *Conversation) ChatCompletionSize(question string, maxTokens int) (openai.FinishReason, error) {
	client, err := GetClient()
	if err != nil {
		return "", err
	}
	ctx := context.Background()

	newQuestion := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	}

	req := openai.ChatCompletionRequest{
		Model:     conv.Model,
		MaxTokens: maxTokens,
		Messages:  append(conv.Messages, newQuestion),
		Stream:    false,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	// TODO : In which case it should not store the question and the answer ?
	conv.Messages = append(conv.Messages, newQuestion)
	conv.Messages = append(conv.Messages, resp.Choices[0].Message)
	conv.HasChange = true

	// Todo : is there a choices 0 in case of err ?
	// TODO : we add the messages to the struct but don't return it
	return resp.Choices[0].FinishReason, err
}

func (conv *Conversation) ChatCompletion(question string) (openai.FinishReason, error) {
	return conv.ChatCompletionSize(question, MaxTokens)
}
