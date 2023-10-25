package main

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"os"
	"sync"
)

type Key struct {
	key   string
	mutex sync.Mutex
}

var key = Key{}

func getKey() (string, error) {
	key.mutex.Lock()
	defer key.mutex.Unlock()
	if key.key == "" {
		dat, err := os.ReadFile("key")
		if err != nil {
			return "", err
		}
		key.key = string(dat)
		fmt.Print("Key loaded") // TODO : remove key
	}
	return key.key, nil
}

// TODO : invalid key on every error or need change
func (key *Key) invalid() bool {
	key.mutex.Lock()
	defer key.mutex.Unlock()
	ok := true
	if key.key == "" {
		ok = false
	}
	key.key = ""
	return ok
}

type OpenClient struct {
	client *openai.Client
	mutex  sync.Mutex
}

var openClient = OpenClient{}

func getClient() (*openai.Client, error) {
	openClient.mutex.Lock()
	defer openClient.mutex.Unlock()
	if openClient.client == nil {
		key, err := getKey()
		if err != nil {
			return nil, err
		}
		openClient.client = openai.NewClient(key)
	}
	return openClient.client, nil
}

func (openClient *OpenClient) invalid() bool {
	openClient.mutex.Lock()
	defer openClient.mutex.Unlock()
	ok := true
	if openClient.client == nil {
		ok = false
	}
	openClient.client = nil
	return ok
}

func (mes *Conversation) chatCompletion(question string, model string) (openai.FinishReason, error) {
	client, err := getClient()
	if err != nil {
		return "", err
	}
	ctx := context.Background()

	newQuestion := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	}
	mes.messages = append(mes.messages, newQuestion)
	mes.hasChange = true

	req := openai.ChatCompletionRequest{
		Model:     model,
		MaxTokens: 20, // TODO : edit hyper parameters
		Messages:  mes.messages,
		Stream:    false,
	}

	resp, err := client.CreateChatCompletion(ctx, req)

	// Todo : is there a choices 0 in case of err ?
	// TODO : we add the messages to the struct but don't return it
	mes.messages = append(mes.messages, resp.Choices[0].Message)
	return resp.Choices[0].FinishReason, err
}
