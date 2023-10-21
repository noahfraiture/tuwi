package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/couchbase/gocb/v2"
	"github.com/sashabaranov/go-openai"
	"os"
	"sync"
	"time"
)

type couchDB struct {
	username         string
	password         string
	connectionString string
	bucketName       string
	scopeName        string
	collectionName   string
	cluster          *gocb.Cluster
	bucket           *gocb.Bucket
	scope            *gocb.Scope
	collection       *gocb.Collection
	sync.Mutex
}

func (selfDB *couchDB) getCluster() (*gocb.Cluster, error) {
	selfDB.Lock()
	defer selfDB.Unlock()
	if selfDB.cluster == nil {
		cluster, err := gocb.Connect("couchbase://"+selfDB.connectionString, gocb.ClusterOptions{
			Authenticator: gocb.PasswordAuthenticator{
				Username: selfDB.username,
				Password: selfDB.password,
			},
		})
		if err != nil {
			return nil, err
		}
		selfDB.cluster = cluster
	}
	return selfDB.cluster, nil
}

func (selfDB *couchDB) getBucket() (*gocb.Bucket, error) {
	selfDB.Lock()
	defer selfDB.Unlock()
	if selfDB.bucket == nil {
		cluster, err := selfDB.getCluster()
		if err != nil {
			return nil, err
		}
		bucket := cluster.Bucket(selfDB.bucketName)
		err = bucket.WaitUntilReady(5*time.Second, nil)
		if err != nil {
			return nil, err
		}
		selfDB.bucket = bucket
	}
	return selfDB.bucket, nil
}

func (selfDB *couchDB) getScope() (*gocb.Scope, error) {
	selfDB.Lock()
	defer selfDB.Unlock()
	if selfDB.scope == nil {
		bucket, err := selfDB.getBucket()
		if err != nil {
			return nil, err
		}
		selfDB.scope = bucket.Scope(selfDB.scopeName)
	}
	return selfDB.scope, nil
}

func (selfDB *couchDB) getCollection() (*gocb.Collection, error) {
	selfDB.Lock()
	defer selfDB.Unlock()
	if selfDB.collection == nil {
		scope, err := selfDB.getScope()
		if err != nil {
			return nil, err
		}
		selfDB.collection = scope.Collection(selfDB.collectionName)
	}
	return selfDB.collection, nil
}

// TODO : if I want to divide my db in multiple scope, I can divide this structure in multiple struct
var db = couchDB{
	username:         "admin",
	password:         "admin123",
	connectionString: "localhost",
	bucketName:       "conversations",
	scopeName:        "_default",
	collectionName:   "_default",
}

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
