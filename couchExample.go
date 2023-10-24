package main

import (
	"fmt"
	"github.com/couchbase/gocb/v2"
	"log"
	"time"
)

func main() {
	// Update this to your cluster details
	connectionString := "localhost"
	bucketName := "conversations"
	username := "admin"
	password := "admin123"

	// For a secure cluster connection, use `couchbase://<your-cluster-ip>` instead.
	cluster, err := gocb.Connect("couchbase://"+connectionString, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	bucket := cluster.Bucket(bucketName)
	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		log.Fatal(err)
	}

	col := bucket.Scope("_default").Collection("_default")

	// Create and store a Document
	type User struct {
		Name      string   `json:"name"`
		Email     string   `json:"email"`
		Interests []string `json:"interests"`
	}

	_, err = col.Upsert("u:jade",
		User{
			Name:      "Jade",
			Email:     "jade@test-email.com",
			Interests: []string{"Swimming", "Rowing"},
		}, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Get the document back
	getResult, err := col.Get("u:jade", nil)
	if err != nil {
		log.Fatal(err)
	}

	var inUser User
	err = getResult.Content(&inUser)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User: %v\n", inUser)
}
