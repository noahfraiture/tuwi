package main

import (
	"fmt"
	"github.com/couchbase/gocb/v2"
	"github.com/sashabaranov/go-openai"
	"sync"
	"time"
)

// TODO : if I want to divide my db in multiple scope, I can divide this structure in multiple struct
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

func (selfDB *couchDB) invalidCollection() {
	selfDB.collection = nil
}

func (selfDB *couchDB) invalidScope() {
	selfDB.invalidCollection()
	selfDB.scope = nil
}

func (selfDB *couchDB) invalidBucket() {
	selfDB.invalidScope()
	selfDB.bucket = nil
}

func (selfDB *couchDB) invalidCluster() error {
	selfDB.invalidBucket()
	err := selfDB.cluster.Close(nil)
	if err != nil {
		return err
	}
	selfDB.cluster = nil
	return nil
}

// Lazy
func (selfDB *couchDB) changeBucket(newBucket string) {
	selfDB.Lock()
	defer selfDB.Unlock()
	selfDB.invalidBucket()
	selfDB.bucketName = newBucket
}

// Lazy
func (selfDB *couchDB) changeScope(newScope string) {
	selfDB.Lock()
	defer selfDB.Unlock()
	selfDB.invalidScope()
	selfDB.scopeName = newScope
}

// Lazy
func (selfDB *couchDB) changeCollection(newCollection string) {
	selfDB.Lock()
	defer selfDB.Unlock()
	selfDB.invalidCollection()
	selfDB.collectionName = newCollection
}

type conversation struct {
	id          string
	length      int
	model       string
	name        string
	messages    []openai.ChatCompletionMessage // TODO : are these message able to be json ?
	hasChange   bool
	*sync.Mutex // TODO : how will it be store in json ?
}

func (selfDB *couchDB) getConversation(id string) (conversation, error) {
	col, err := selfDB.getCollection()
	if err != nil {
		return conversation{}, nil
	}
	res, err := col.Get(id, nil) // TODO : check possible option and api
	if err != nil {
		return conversation{}, err
	}
	conv := conversation{}
	err = res.Content(&conv)
	return conv, err
}

func (selfDB *couchDB) storeConversation(conv *conversation) error {
	// TODO : upsert or insert ?
	col, err := selfDB.getCollection()
	if err != nil {
		return err
	}
	_, err = col.Upsert(conv.id, conv, nil) // TODO : what about mutation returned value
	return err
}

// TODO : couldn't I get only the id to avoid useless call ?
func (selfDB *couchDB) getDocumentsID() ([]string, error) {
	scope, err := selfDB.getScope()
	if err != nil {
		return nil, err
	}
	queryRes, err := scope.Query(
		fmt.Sprintf("SELECT META().id FROM %s", selfDB.collectionName),
		&gocb.QueryOptions{},
	)
	if err != nil {
		return nil, err
	}
	idList := make([]string, 8)
	for queryRes.Next() {
		var tmpID string
		err = queryRes.Row(&tmpID)
		if err != nil {
			return nil, err
		}
		idList = append(idList, tmpID)
	}
	return idList, nil
}
