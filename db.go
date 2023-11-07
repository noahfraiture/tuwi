package main

import (
	"errors"
	"fmt"
	"github.com/couchbase/gocb/v2"
	"github.com/sashabaranov/go-openai"
	"time"
)

type (
	// Conversation TODO : Where should I put this shit ?
	Conversation struct {
		ID        string                         `json:"id"`
		Model     string                         `json:"model"`
		Name      string                         `json:"name"`
		Messages  []openai.ChatCompletionMessage `json:"messages"` // TODO : are these message able to be json ?
		HasChange bool                           `json:"has_change"`
	}

	// CouchDB TODO : if I want to divide my db in multiple scope, I can divide this structure in multiple struct
	CouchDB struct {
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
	}
)

func (selfDB *CouchDB) getCluster() (*gocb.Cluster, error) {
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

func (selfDB *CouchDB) getBucket() (*gocb.Bucket, error) {
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

func (selfDB *CouchDB) getScope() (*gocb.Scope, error) {
	if selfDB.scope == nil {
		bucket, err := selfDB.getBucket()
		if err != nil {
			return nil, err
		}
		selfDB.scope = bucket.Scope(selfDB.scopeName)
	}
	return selfDB.scope, nil
}

func (selfDB *CouchDB) getCollection() (*gocb.Collection, error) {
	if selfDB.collection == nil {
		scope, err := selfDB.getScope()
		if err != nil {
			return nil, err
		}
		selfDB.collection = scope.Collection(selfDB.collectionName)
		// TODO : see how to handle non valid collection name
	}
	return selfDB.collection, nil
}

func (selfDB *CouchDB) Get(element string) (any, error) {
	switch element {
	case "collection":
		return selfDB.getCollection()
	case "scope":
		return selfDB.getScope()
	case "bucket":
		return selfDB.getBucket()
	case "cluster":
		return selfDB.getCluster()
	default:
		return nil, errors.New("wrong element in argument")
	}
}

func (selfDB *CouchDB) invalidCollection() {
	selfDB.collection = nil
}

func (selfDB *CouchDB) invalidScope() {
	selfDB.invalidCollection()
	selfDB.scope = nil
}

func (selfDB *CouchDB) invalidBucket() {
	selfDB.invalidScope()
	selfDB.bucket = nil
}

func (selfDB *CouchDB) invalidCluster() error {
	selfDB.invalidBucket()
	err := selfDB.cluster.Close(nil)
	if err != nil {
		return err
	}
	selfDB.cluster = nil
	return nil
}

// Invalid TODO : not fan of element as a string
func (selfDB *CouchDB) Invalid(element string) error {
	switch element {
	case "cluster":
		err := selfDB.invalidCluster()
		if err != nil {
			return err
		}
	case "bucket":
		selfDB.invalidBucket()
	case "scope":
		selfDB.invalidScope()
	case "collection":
		selfDB.invalidCollection()
	default:
		return errors.New("wrong element in argument")
	}
	return nil
}

// ChangeBucket Lazy
func (selfDB *CouchDB) ChangeBucket(newBucket string) bool {
	if selfDB.bucketName != newBucket {
		selfDB.invalidBucket()
		selfDB.bucketName = newBucket
		return true
	}
	return false
}

// ChangeScope Lazy
func (selfDB *CouchDB) ChangeScope(newScope string) bool {
	if selfDB.scopeName != newScope {
		selfDB.invalidScope()
		selfDB.scopeName = newScope
		return true
	}
	return false
}

// ChangeCollection Lazy
func (selfDB *CouchDB) ChangeCollection(newCollection string) bool {
	if selfDB.collectionName != newCollection {
		selfDB.invalidCollection()
		selfDB.collectionName = newCollection
		return true
	}
	return false
}

func (selfDB *CouchDB) GetConversation(id string) (Conversation, error) {
	col, err := selfDB.getCollection()
	if err != nil {
		return Conversation{}, nil
	}
	res, err := col.Get(id, nil) // TODO : check possible option and api
	if err != nil {
		return Conversation{}, err
	}
	conv := Conversation{}
	err = res.Content(&conv)
	if conv.HasChange {
		conv.HasChange = false
		err = errors.Join(err, errors.New("conversation has change value to true"))
	}
	return conv, err
}

func (selfDB *CouchDB) StoreConversation(conv *Conversation) error {
	// TODO : upsert or insert ?
	col, err := selfDB.getCollection()
	if err != nil {
		return err
	}
	_, err = col.Upsert(conv.ID, conv, nil)
	conv.HasChange = false
	return err
}

func (selfDB *CouchDB) GetDocumentsID() ([]string, error) {
	scope, err := selfDB.getScope()
	if err != nil {
		return nil, err
	}
	queryRes, err := scope.Query(
		fmt.Sprintf("SELECT META().id FROM %s", selfDB.collectionName),
		&gocb.QueryOptions{},
	)
	defer func(queryRes *gocb.QueryResult) {
		err = queryRes.Close()
	}(queryRes)
	if err != nil {
		return nil, err
	}

	idList := make([]string, 0)
	for queryRes.Next() {
		var tmpID struct {
			ID string `json:"id"`
		}
		err = queryRes.Row(&tmpID)
		if err != nil {
			return nil, err
		}
		idList = append(idList, tmpID.ID)
	}
	return idList, nil
}

// TODO : Change to query directly everything and not by id
func (selfDB *CouchDB) GetConversations() ([]Conversation, error) {
	idList, err := selfDB.GetDocumentsID()
	if err != nil {
		return nil, err
	}
	convList := make([]Conversation, len(idList))
	for i, id := range idList {
		convList[i], err = selfDB.GetConversation(id)
		if err != nil {
			return nil, err
		}
	}
	return convList, nil
}
