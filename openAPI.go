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

func validKey(key string) bool {
	regex := regexp.MustCompile(`^sk-[a-zA-Z0-9]{48}$`)
	return regex.MatchString(key)
}

func createKey(key string) error {
	return os.WriteFile("key", []byte(key), 0644)
}

// Lazy load
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
	if !validKey(string(data)) {
		return "", errors.New("key is invalid")
	}
	tmpKey := strings.TrimRight(string(data), "\n")
	if !validKey(tmpKey) {
		return "", errors.New("key is invalid")
	}

	// Key file exist and is valid
	key = Key(tmpKey)
	return key, nil
}

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

func (conv *Conversation) chatCompletionSizeModel(maxTokens int, model string, c chan gptMessage) {
	// TODO : handle error
	client, _ := GetClient()
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages:  conv.openaiMessages(), // Note : This already contains the question
		Stream:    false,
	}

	resp, _ := client.CreateChatCompletion(ctx, req)
	c <- gptMessage(resp.Choices[0])
	close(c)
}

func (conv *Conversation) chatCompletionModel(model string, c chan gptMessage) {
	conv.chatCompletionSizeModel(MaxTokens, model, c)
}

func (conv *Conversation) chatCompletionSize(maxTokens int, c chan gptMessage) {
	conv.chatCompletionSizeModel(maxTokens, conv.LastModel, c)
}

func (conv *Conversation) chatCompletion(c chan gptMessage) {
	conv.chatCompletionModel(conv.LastModel, c)
}
