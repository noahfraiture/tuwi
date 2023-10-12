package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
	"sync"
)

type conversation struct {
	length   int
	model    string
	name     string
	messages []openai.ChatCompletionMessage
	valid    bool
	*sync.Mutex
}

func (c conversation) invalid() {
	c.Lock()
	defer c.Unlock()
	c.messages = nil
	c.valid = false
}

// TODO : test pointer
func store(conv conversation) error {
	conv.Lock()
	defer conv.Unlock()
	if !conv.valid {
		return errors.New("conversation is invalid, can't store")
	}
	jsonBytes, err := json.Marshal(conv) // TODO : will store mutex ?
	if err != nil {
		return err
	}

	f, err := os.OpenFile("/db/"+conv.name, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(f)
	_, err = f.Write(jsonBytes)
	if err != nil {
		return err
	}
	return nil
}

func loadConversation(filename string) (conversation, error) {
	var conv conversation

	jsonBytes, err := os.ReadFile(filename)
	if err != nil {
		return conv, err // todo : handle errors
	}

	err = json.Unmarshal(jsonBytes, &conv)
	if err != nil {
		return conv, err
	}
	// TODO : check if value is well always true and if not handle the case since it can't be load better
	return conv, nil
}

type history struct {
	valid   bool
	history []string
	*sync.Mutex
}

func loadHistory() (history, error) {
	var hist history

	jsonBytes, err := os.ReadFile("history.json")
	if err != nil {
		return hist, err
	}

	err = json.Unmarshal(jsonBytes, &hist)
	if err != nil {
		return hist, err
	}
	return hist, nil
}
