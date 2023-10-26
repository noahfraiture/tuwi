package main

import (
	"context"
	openai "github.com/sashabaranov/go-openai"
	"os"
	"sync"
)

type Key struct {
	Key string
	sync.Mutex
}

var key = Key{}

func GetKey() (string, error) {
	key.Lock()
	defer key.Unlock()
	if key.Key == "" {
		dat, err := os.ReadFile("key")
		if err != nil {
			return "", err
		}
		key.Key = string(dat)
	}
	return key.Key, nil
}

// Invalid TODO : Invalid key on every error or need change
// Should never be used outside a function with a malloc
func (key *Key) Invalid() bool {
	if key.Key == "" {
		return false
	}
	key.Key = ""
	return true
}

type OpenClient struct {
	client *openai.Client
	mutex  sync.Mutex
}

var openClient = OpenClient{}

func GetClient() (*openai.Client, error) {
	openClient.mutex.Lock()
	defer openClient.mutex.Unlock()
	if openClient.client == nil {
		key, err := GetKey()
		if err != nil {
			return nil, err
		}
		openClient.client = openai.NewClient(key)
	}
	return openClient.client, nil
}

func (openClient *OpenClient) Invalid() bool {
	openClient.mutex.Lock()
	defer openClient.mutex.Unlock()
	ok := true
	if openClient.client == nil {
		ok = false
	}
	openClient.client = nil
	return ok
}

func (selfDoc *Conversation) ChatCompletion(question string) (openai.FinishReason, error) {
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
		Model:     selfDoc.Model,
		MaxTokens: 20, // TODO : edit hyper parameters
		Messages:  append(selfDoc.Messages, newQuestion),
		Stream:    false,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	// TODO : In which case it should not store the question and the answer ?
	selfDoc.Messages = append(selfDoc.Messages, newQuestion)
	selfDoc.Messages = append(selfDoc.Messages, resp.Choices[0].Message)
	selfDoc.HasChange = true

	// Todo : is there a choices 0 in case of err ?
	// TODO : we add the messages to the struct but don't return it
	return resp.Choices[0].FinishReason, err
}
