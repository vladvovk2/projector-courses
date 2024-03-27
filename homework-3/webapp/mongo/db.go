package mongo

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client *mongo.Client
)

// Init initializes MongoDB connection
func Init(uri string) error {
	// Set up MongoDB client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the MongoDB server to ensure that the connection was successful
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	log.Println("Connected to MongoDB")
	return nil
}

// GetClient returns the MongoDB client
func GetClient() *mongo.Client {
	return client
}
