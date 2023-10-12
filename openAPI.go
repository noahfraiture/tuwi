package main

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"os"
	"sync"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Key struct {
	key   string
	mutex sync.Mutex
}

var key = Key{}

func getKey() string {
	key.mutex.Lock()
	defer key.mutex.Unlock()
	if key.key == "" {
		dat, err := os.ReadFile("key")
		check(err)
		key.key = string(dat)
		fmt.Print("Key loaded") // TODO : remove key
	}
	return key.key
}

type OpenClient struct {
	client *openai.Client
	mutex  sync.Mutex
}

var openClient = OpenClient{}

func getClient() *openai.Client {
	openClient.mutex.Lock()
	defer openClient.mutex.Unlock()
	if openClient.client == nil {
		openClient.client = openai.NewClient(getKey())
	}
	return openClient.client
}

// Todo : gotta decide if I do procedure or function.
// Procedure make it un-natural to create alternative. I gotta see if it's a real thing.
// If it's rare thing it may not be interesting and cost too much
// For now it's gonna be procedure for this and function for new conv
func chatCompletion(question string, messages *[]openai.ChatCompletionMessage, model string) (openai.FinishReason, error) {
	client := getClient()
	ctx := context.Background()

	newMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	}
	*messages = append(*messages, newMessage)

	req := openai.ChatCompletionRequest{
		Model:     model,
		MaxTokens: 20, // TODO : edit hyper parameters
		Messages:  *messages,
		Stream:    false,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	return resp.Choices[0].FinishReason, err // Todo : is there a choices 0 in case of err ?
}

func newConversation(question string, model string, systemMessage string) []openai.ChatCompletionMessage {
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}
	var messages []openai.ChatCompletionMessage
	if systemMessage != "" {
		messages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMessage,
			},
		}
	} else {
		messages = []openai.ChatCompletionMessage{}
	}
	_, err := chatCompletion(question, &messages, model)
	// todo : handle finished reason and errors
	if err != nil {
		fmt.Println(err)
	}
	return messages
}
