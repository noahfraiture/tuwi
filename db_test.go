package main

import (
	"errors"
	"github.com/couchbase/gocb/v2"
	"github.com/sashabaranov/go-openai"
	"reflect"
	"sync"
	"testing"
)

type (
	docMut struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		Number   float64 `json:"number"`
		NotInDoc string  `json:"notInDoc"`
		sync.Mutex
	}

	docSimp struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		Number   float64 `json:"number"`
		NotInDoc string  `json:"notInDoc"` // Specify the key in json
	}

	document interface {
		getID() string
		getName() string
		getNumber() float64
		getNotInDoc() string
	}
)

var (
	dbTest = CouchDB{
		username:         "admin",
		password:         "admin123",
		connectionString: "localhost",
		bucketName:       "test",
		scopeName:        "_default",
		collectionName:   "_default",
		cluster:          nil,
		bucket:           nil,
		scope:            nil,
		collection:       nil,
	}

	m0 document = &docMut{
		ID:       "m0",
		Name:     "Jon",
		Number:   19,
		NotInDoc: "Should not be in doc",
	}
	m1 document = &docMut{
		ID:   "m1",
		Name: "Janne",
	}
	m2 document = &docMut{
		ID:       "m2",
		Number:   20,
		NotInDoc: "Should not be in doc",
		Mutex:    sync.Mutex{},
	}

	s0 document = &docSimp{
		ID:       "s0",
		Name:     "Jon",
		Number:   19,
		NotInDoc: "Should not be in doc",
	}
	s1 document = &docSimp{
		ID:   "s1",
		Name: "Janne",
	}
	s2 document = &docSimp{
		ID:       "s2",
		Number:   20,
		NotInDoc: "Should not be in doc",
	}

	listID = []string{"m0", "m1", "m2", "s0", "s1", "s2"}

	c0 document = &Conversation{
		ID:        "c0",
		Model:     openai.GPT3Dot5Turbo,
		Name:      "jon",
		Messages:  nil,
		HasChange: false,
	}

	c1 document = &Conversation{
		ID:        "c1",
		Model:     openai.GPT3Dot5Turbo,
		Name:      "janne",
		Messages:  nil,
		HasChange: true,
	}

	c2 document = &Conversation{
		ID:    "c2",
		Model: openai.GPT3Dot5Turbo,
		Name:  "jon",
		Messages: []openai.ChatCompletionMessage{
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

func (selfDoc *docMut) getID() string {
	return selfDoc.ID
}

func (selfDoc *docSimp) getID() string {
	return selfDoc.ID
}

func (conv *Conversation) getID() string {
	return conv.ID
}

func (selfDoc *docMut) getName() string {
	return selfDoc.Name
}

func (selfDoc *docSimp) getName() string {
	return selfDoc.Name
}

func (conv *Conversation) getName() string {
	return conv.Name
}

func (selfDoc *docMut) getNumber() float64 {
	return selfDoc.Number
}

func (selfDoc *docSimp) getNumber() float64 {
	return selfDoc.Number
}

func (conv *Conversation) getNumber() float64 {
	return 0
}

func (selfDoc *docMut) getNotInDoc() string {
	return selfDoc.NotInDoc
}

func (selfDoc *docSimp) getNotInDoc() string {
	return selfDoc.NotInDoc
}

func (conv *Conversation) getNotInDoc() string {
	return ""
}

// Test once in case it doesn't exist, and once in case it already does
func TestCouchDB_GetCluster(t *testing.T) {
	cluster, err := dbTest.getCluster()
	if err != nil {
		t.Error(err)
	}
	if cluster == nil {
		t.Error("cluster is nil at first call")
	}

	cluster, err = dbTest.getCluster()
	if err != nil {
		t.Error(err)
	}
	if cluster == nil {
		t.Error("cluster is nil at second call")
	}
}

// Test once in case it doesn't exist, and once in case it already does
func TestCouchDB_GetBucket(t *testing.T) {
	bucket, err := dbTest.getBucket()
	if err != nil {
		t.Error(err)
	}
	if bucket == nil {
		t.Error("bucket is nil at first call")
	}

	bucket, err = dbTest.getBucket()
	if err != nil {
		t.Error(err)
	}
	if bucket == nil {
		t.Error("bucket is nil at second call")
	}
}

// Test once in case it doesn't exist, and once in case it already does
func TestCouchDB_GetScope(t *testing.T) {
	scope, err := dbTest.getScope()
	if err != nil {
		t.Error(err)
	}
	if scope == nil {
		t.Error("scope is nil at first call")
	}

	scope, err = dbTest.getScope()
	if err != nil {
		t.Error(err)
	}
	if scope == nil {
		t.Error("scope is nil at second call")
	}
}

// Test once in case it doesn't exist, and once in case it already does
func TestCouchDB_GetCollection(t *testing.T) {
	collection, err := dbTest.getCollection()
	if err != nil {
		t.Error(err)
	}
	if collection == nil {
		t.Error("collection is nil at first call")
	}

	collection, err = dbTest.getCollection()
	if err != nil {
		t.Error(err)
	}
	if collection == nil {
		t.Error("collection is nil at second call")
	}
}

func TestCouchDB_InvalidCollection(t *testing.T) {
	_, err := dbTest.Get("collection")
	if err != nil {
		t.Error(err)
	}
	err = dbTest.Invalid("collection")
	if err != nil {
		t.Error(err)
	}
	if dbTest.collection != nil {
		t.Error("collection is not nil after invalidation")
	}
}

func TestCouchDB_InvalidScope(t *testing.T) {
	_, err := dbTest.Get("scope")
	if err != nil {
		t.Error(err)
	}
	err = dbTest.Invalid("scope")
	if err != nil {
		t.Error(err)
	}
	if dbTest.scope != nil {
		t.Error("scope is not nil after invalidation")
	}
}

func TestCouchDB_InvalidBucket(t *testing.T) {
	_, err := dbTest.Get("bucket")
	if err != nil {
		t.Error(err)
	}
	err = dbTest.Invalid("bucket")
	if err != nil {
		t.Error(err)
	}
	if dbTest.bucket != nil {
		t.Error("bucket is not nil after invalidation")
	}
}

func TestCouchDB_InvalidCluster(t *testing.T) {
	_, err := dbTest.Get("cluster")
	if err != nil {
		t.Error(err)
	}
	err = dbTest.Invalid("cluster")
	if err != nil {
		t.Error(err)
	}
	if dbTest.cluster != nil {
		t.Error("cluster is not nil after invalidation")
	}
}

func addDoc(docs ...document) error {
	col, err := dbTest.getCollection()
	if err != nil {
		return err
	}
	for _, doc := range docs {
		_, err = col.Upsert(doc.getID(), &doc, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func readDocs(docs ...document) error {
	col, err := dbTest.getCollection()
	if err != nil {
		return err
	}
	savedDocs := make([]any, len(docs))
	for i, doc := range docs {
		var res *gocb.GetResult
		res, err = col.Get(doc.getID(), nil)
		if err != nil {
			return err
		}
		err = res.Content(&savedDocs[i])
		if err != nil {
			return err
		}
		inter := savedDocs[i].(map[string]interface{})
		switch {
		case doc.getID() != inter["id"].(string):
			return errors.New("id is not the same")
		case doc.getName() != inter["name"].(string):
			return errors.New("name is not the same")
		case doc.getNumber() != inter["number"].(float64):
			return errors.New("number is not the same")
		case doc.getNotInDoc() != inter["notInDoc"].(string):
			return errors.New("notInDoc is not the same")
		}
	}
	return nil
}

func TestCouchDB_AddDocSimp(t *testing.T) {
	err := addDoc(s0, s1, s2)
	if err != nil {
		t.Error(err)
	}
}

func TestCouchDB_ReadDocSimp(t *testing.T) {
	err := readDocs(s0, s1, s2)
	if err != nil {
		t.Error(err)
	}
}

func TestCouchDB_AddDocMut(t *testing.T) {
	err := addDoc(m0, m1, m2)
	if err != nil {
		t.Error(err)
	}
}

func TestCouchDB_ReadDocMut(t *testing.T) {
	err := readDocs(m0, m1, m2)
	if err != nil {
		t.Error(err)
	}
}

// In bucket test after adding documents
func TestCouchDB_GetDocumentsID(t *testing.T) {
	ids, err := dbTest.GetDocumentsID()
	if err != nil {
		t.Error(err)
	}
	if ids == nil {
		t.Error("ids is nil")
	}
	if !reflect.DeepEqual(ids, listID) {
		t.Error("ids is not equal to listStringTest")
	}
}

// Change for test2
func TestCouchDB_ChangeBucket(t *testing.T) {
	_, err := dbTest.getCollection()
	if err != nil {
		t.Error(err)
	}
	hasChanged := dbTest.ChangeBucket("test2")
	if !hasChanged {
		t.Error("hasChanged is false after change")
	}
	if dbTest.bucketName != "test2" {
		t.Error("bucketName is not test2 after change")
	}
}

func TestCouchDB_AddConv(t *testing.T) {
	err := addDoc(c0, c1, c2)
	if err != nil {
		t.Error(err)
	}
}

// Test for conv are not necessary, hand works

func Test_Clear(t *testing.T) {
	dbTest.ChangeBucket("test")
	col, err := dbTest.getCollection()
	if err != nil {
		t.Error(err)
	}
	ids, err := dbTest.GetDocumentsID()
	if err != nil {
		t.Error(err)
	}
	for _, id := range ids {
		_, err = col.Remove(id, nil)
		if err != nil {
			t.Error(err)
		}
	}
	dbTest.ChangeBucket("test2")
	col, err = dbTest.getCollection()
	if err != nil {
		t.Error(err)
	}
	ids, err = dbTest.GetDocumentsID()
	if err != nil {
		t.Error(err)
	}
	for _, id := range ids {
		_, err = col.Remove(id, nil)
		if err != nil {
			t.Error(err)
		}
	}
}
