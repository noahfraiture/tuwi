package main

import (
	"encoding/json"
	"os"
)

const dbPath = "./db/"

type Conversations map[string]Conversation

func readConversation(id string) (Conversation, error) {
	jsonFile, err := os.ReadFile(dbPath + id + ".json")
	if err != nil {
		return Conversation{}, err
	}

	conv := Conversation{}
	err = json.Unmarshal(jsonFile, &conv)
	conv.HasChange = true
	return conv, err
}

func clearDB() error {
	ids, err := getIDS()
	if err != nil {
		return err
	}
	for _, id := range ids {
		err = os.Remove(dbPath + id + ".json")
		if err != nil {
			return err
		}
	}
	return nil
}

func getIDS() ([]string, error) {
	files, err := os.ReadDir(dbPath + ".")
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ids = append(ids, file.Name()[:len(file.Name())-5])
	}
	return ids, nil
}

func (conversations *Conversations) UpdateConversations() error {
	ids, err := getIDS()
	if err != nil {
		return err
	}
	for _, id := range ids {
		conv, ok := (*conversations)[id]
		if !ok || conv.HasChange {
			conv, err = readConversation(id)
			if err != nil {
				return err
			}
			(*conversations)[id] = conv
		}
	}
	return nil
}

func (conversations *Conversations) GetConversation(id string) (Conversation, error) {
	conv, ok := (*conversations)[id]
	if !ok || conv.HasChange {
		var err error
		conv, err = readConversation(id)
		if err != nil {
			return Conversation{}, err
		}
		(*conversations)[id] = conv
	}
	return conv, nil
}

func (conv *Conversation) SaveConversation() error {
	conv.HasChange = false
	file := dbPath + conv.ID + ".json"
	jsonConv, err := json.Marshal(conv)
	if err != nil {
		conv.HasChange = true
		return err
	}
	err = os.WriteFile(file, jsonConv, 0644)
	if err != nil {
		conv.HasChange = true
		return err
	}
	return nil
}

func (conv *Conversation) Invalid() {
	conv.HasChange = true
}
