package main

import (
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"testing"
)

func testGetKey() error {
	content, err := getKey()
	if err != nil {
		return err
	}
	if content == "" {
		return errors.New("key is empty")
	}
	return nil
}

func Test_GetKey(t *testing.T) {
	if err := testGetKey(); err != nil {
		t.Error(err)
	}
	if err := testGetKey(); err != nil {
		t.Error(err)
	}
}

func TestKey_Invalid(t *testing.T) {
	if res := key.invalid(); !res {
		t.Error("The key was empty before invalidating")
	}
	if res := key.invalid(); res {
		t.Error("The key plain after an invalidation")
	}
	if string(key) != "" {
		t.Error("How tf is it not empty now ?")
	}
}

func testGetClient() error {
	content, err := GetClient()
	if err != nil {
		return err
	}
	if content == nil {
		return errors.New("client is nil")
	}
	return nil
}

func Test_GetClient(t *testing.T) {
	if err := testGetClient(); err != nil {
		t.Error(err)
	}
	if err := testGetClient(); err != nil {
		t.Error(err)
	}
}

func TestOpenClient_Invalid(t *testing.T) {
	if res := openClient.invalid(); !res {
		t.Error("The client was empty before invalidating")
	}
	if res := openClient.invalid(); res {
		t.Error("The client plain after an invalidation")
	}
	if openClient.client != nil {
		t.Error("How tf is it not empty now ?")
	}
}

func TestConversation_ChatCompletion_Empty(t *testing.T) {
	choice := openai.ChatCompletionChoice{
		Index: 0,
		Message: openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are an helpful assistant",
		},
		FinishReason: "",
	}
	conversation := Conversation{
		ID:        "conv1",
		LastModel: openai.GPT3Dot5Turbo,
		Name:      "Conversation 1",
		Messages:  []Message{gptMessage(choice).toMessage(openai.GPT3Dot5Turbo)},
		HasChange: false,
	}
	question := Message{
		Role:         roleUser,
		Content:      "Hello",
		FinishReason: finishUser,
		Model:        openai.GPT3Dot5Turbo,
	}
	conversation.Messages = append(conversation.Messages, question)
	err := conversation.chatCompletion()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("The response : %s\n", conversation.Messages[len(conversation.Messages)-1].Content)
	if conversation.Messages[len(conversation.Messages)-1].FinishReason != finishReason(openai.FinishReasonStop) {
		t.Errorf("The conversation should have stopped : %s", conversation.Messages[len(conversation.Messages)-1].FinishReason)
	}
	if len(conversation.Messages) != 3 {
		t.Error("The conversation should have 3 messages but have ", len(conversation.Messages))
	}
	if conversation.HasChange != true {
		t.Error("The conversation should have change flag")
	}
}

func TestConversation_ChatCompletion_Full(t *testing.T) {
	choices := []openai.ChatCompletionChoice{
		{
			Message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a cool friend",
			},
		},
		{
			Message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hey",
			},
		},
		{
			Message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Yo",
			},
		},
	}
	messages := make([]Message, 3, 5)
	for i, choice := range choices {
		messages[i] = gptMessage(choice).toMessage(openai.GPT3Dot5Turbo)
	}
	conversation := Conversation{
		ID:        "conv2",
		LastModel: openai.GPT3Dot5Turbo,
		Name:      "Conversation 2",
		Messages:  messages,
		HasChange: false,
	}
	question := Message{
		Role:         roleUser,
		Content:      "Say 'banana'",
		FinishReason: finishUser,
		Model:        openai.GPT3Dot5Turbo,
	}
	conversation.Messages = append(conversation.Messages, question)
	err := conversation.chatCompletion()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("The response : %s\n", conversation.Messages[len(conversation.Messages)-1].Content)
	if conversation.Messages[len(conversation.Messages)-1].FinishReason != finishReason(openai.FinishReasonStop) {
		t.Errorf("The conversation should have stopped : %s", conversation.Messages[len(conversation.Messages)-1].FinishReason)
	}
	if len(conversation.Messages) != 5 {
		t.Error("The conversation should have 5 messages")
	}
	if conversation.HasChange != true {
		t.Error("The conversation should have change flag")
	}
}

func TestConversation_ChatCompletion_TooLong(t *testing.T) {
	choices := []openai.ChatCompletionChoice{
		{
			Message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a cool friend",
			},
		},
		{
			Message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hey",
			},
		},
		{
			Message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Yo",
			},
		},
	}
	messages := make([]Message, 3, 5)
	for i, choice := range choices {
		messages[i] = gptMessage(choice).toMessage(openai.GPT3Dot5Turbo)
	}
	conversation := Conversation{
		ID:        "conv3",
		LastModel: openai.GPT3Dot5Turbo,
		Name:      "Conversation 3",
		Messages:  messages,
		HasChange: false,
	}
	question := Message{
		Role:         roleUser,
		Content:      "how do you do ? I do fine for my self, I think the most important thing here is that you feel good too. I would like you to explain to me in a few page how you feel",
		FinishReason: finishUser,
		Model:        openai.GPT3Dot5Turbo,
	}
	conversation.Messages = append(conversation.Messages, question)
	err := conversation.chatCompletionSize(10)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("The response : %s\n", conversation.Messages[len(conversation.Messages)-1].Content)
	if conversation.Messages[len(conversation.Messages)-1].FinishReason != finishReason(openai.FinishReasonLength) {
		t.Errorf("The conversation should have stopped : %s", conversation.Messages[len(conversation.Messages)-1].FinishReason)
	}
	if len(conversation.Messages) != 5 {
		t.Error("The conversation should have 5 messages")
	}
	if conversation.HasChange != true {
		t.Error("The conversation should not have change flag")
	}
}
