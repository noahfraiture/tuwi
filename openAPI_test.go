package main

import (
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"testing"
)

func testGetKey() error {
	content, err := GetKey()
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
	if res := key.Invalid(); !res {
		t.Error("The key was empty before invalidating")
	}
	if res := key.Invalid(); res {
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
	if res := openClient.Invalid(); !res {
		t.Error("The client was empty before invalidating")
	}
	if res := openClient.Invalid(); res {
		t.Error("The client plain after an invalidation")
	}
	if openClient.client != nil {
		t.Error("How tf is it not empty now ?")
	}
}

func TestConversation_ChatCompletion_Empty(t *testing.T) {
	conversation := Conversation{
		ID:    "conv1",
		Model: openai.GPT3Dot5Turbo,
		Name:  "Conversation 1",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem,
				Content: "You are a cool friend"}, // DOESN'T WORK WITH NO MESSAGE. ERROR 'too short'
		},
		HasChange: false,
	}
	finishReason, err := conversation.ChatCompletion("Hello")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("The response : %s\n", conversation.Messages[len(conversation.Messages)-1].Content)
	if finishReason != openai.FinishReasonStop {
		t.Error("The conversation should have stopped")
	}
	if len(conversation.Messages) != 3 {
		t.Error("The conversation should have 3 messages")
	}
	if conversation.HasChange != true {
		t.Error("The conversation should have change flag")
	}
}

func TestConversation_ChatCompletion_Full(t *testing.T) {
	conversation := Conversation{
		ID:    "conv2",
		Model: openai.GPT3Dot5Turbo,
		Name:  "Conversation 2",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem,
				Content: "You are a cool friend"},
			{Role: openai.ChatMessageRoleUser,
				Content: "Hey"},
			{Role: openai.ChatMessageRoleAssistant,
				Content: "Yo"},
		},
		HasChange: false,
	}
	finishReason, err := conversation.ChatCompletion("Say 'banana'")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("The response : %s\n", conversation.Messages[len(conversation.Messages)-1].Content)
	if finishReason != openai.FinishReasonStop {
		t.Errorf("The conversation should have stopped : %s", finishReason)
	}
	if len(conversation.Messages) != 5 {
		t.Error("The conversation should have 5 messages")
	}
	if conversation.HasChange != true {
		t.Error("The conversation should have change flag")
	}
}

func TestConversation_ChatCompletion_TooLong(t *testing.T) {
	conversation := Conversation{
		ID:    "conv3",
		Model: openai.GPT3Dot5Turbo,
		Name:  "Conversation 3",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem,
				Content: "You are a cool friend"},
			{Role: openai.ChatMessageRoleUser,
				Content: "Hey"},
			{Role: openai.ChatMessageRoleAssistant,
				Content: "Yo"},
		},
		HasChange: false,
	}
	finishReason, err := conversation.ChatCompletion("how do you do ? I do fine for my self, " +
		"I think the most important thing here is that you feel good too. " +
		"I would like you to explain to me in a few page how you feel")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("The response : %s\n", conversation.Messages[len(conversation.Messages)-1].Content)
	if finishReason != openai.FinishReasonLength {
		t.Error("The conversation should have stopped")
	}
	if len(conversation.Messages) != 5 {
		t.Error("The conversation should have 5 messages")
	}
	if conversation.HasChange != true {
		t.Error("The conversation should not have change flag")
	}
}
