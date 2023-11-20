package main

import (
	"github.com/sashabaranov/go-openai"
	"testing"
)

var (
	c0 = Conversation{
		ID:        "c0",
		LastModel: openai.GPT3Dot5Turbo,
		Name:      "jon",
		Messages:  nil,
		HasChange: false,
	}

	c1 = Conversation{
		ID:        "c1",
		LastModel: openai.GPT3Dot5Turbo,
		Name:      "janne",
		Messages:  nil,
		HasChange: true,
	}

	c2 = Conversation{
		ID:        "c2",
		LastModel: openai.GPT3Dot5Turbo,
		Name:      "jon",
		Messages: []Message{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "hey",
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "yo",
			},
		},
		HasChange: false,
	}
)

func (conv *Conversation) isEqual(other Conversation) bool {
	if conv.ID != other.ID {
		return false
	}
	if conv.LastModel != other.LastModel {
		return false
	}
	if conv.Name != other.Name {
		return false
	}
	if len(conv.Messages) != len(other.Messages) {
		return false
	}
	for i, message := range conv.Messages {
		if message.Role != other.Messages[i].Role {
			return false
		}
		if message.Content != other.Messages[i].Content {
			return false
		}
	}
	return true
}

func TestConversation_SaveConversationAndRead(t *testing.T) {
	err := clearDB()
	if err != nil {
		t.Error(err)
	}
	err = c0.saveConversation()
	if err != nil {
		t.Error(err)
	}
	err = c1.saveConversation()
	if err != nil {
		t.Error(err)
	}
	err = c2.saveConversation()
	if err != nil {
		t.Error(err)
	}

	conv, err := readConversation(c0.ID)
	if err != nil {
		t.Error(err)
	}
	if !conv.isEqual(c0) {
		t.Error("c0 is not c0")
	}
	conv, err = readConversation(c1.ID)
	if err != nil {
		t.Error(err)
	}
	if !conv.isEqual(c1) {
		t.Error("c1 is not c1")
	}
	conv, err = readConversation(c2.ID)
	if err != nil {
		t.Error(err)
	}
	if !conv.isEqual(c2) {
		t.Error("c2 is not c2")
	}
}

func TestConversations_GetConversation(t *testing.T) {
	err := clearDB()
	if err != nil {
		t.Error(err)
	}
	conversations := make(Conversations)
	err = c0.saveConversation()
	conv, err := conversations.getConversation(c0.ID)
	if err != nil {
		t.Error(err)
	}
	if !conv.isEqual(c0) {
		t.Error("c0 is not c0")
	}
	conv, err = conversations.getConversation(c0.ID)
	if err != nil {
		t.Error(err)
	}
	if !conv.isEqual(c0) {
		t.Error("c0 is not c0")
	}
	if c, ok := conversations[c0.ID]; !ok {
		t.Error("c0 is in conversations")
	} else {
		if !c.isEqual(c0) {
			t.Error("c0 is not c0")
		}
	}
}
