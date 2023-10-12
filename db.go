package main

import (
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
)

type conversation struct {
	length   int
	model    string
	name     string
	messages []openai.ChatCompletionMessage
}

func store(conv conversation) {
	jsonBytes, err := json.Marshal(conv)
	if err != nil {
		fmt.Println(err) // TODO : error handling
		return
	}

	f, err := os.OpenFile("/db/"+conv.name, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
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
		fmt.Println(err)
		return
	}
}

func load(filename string) conversation {
	var conv conversation

	jsonBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return conv // todo : handle errors
	}

	err = json.Unmarshal(jsonBytes, &conv)
	if err != nil {
		fmt.Println(err)
	}
	return conv
}
