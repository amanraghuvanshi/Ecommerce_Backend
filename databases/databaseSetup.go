package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSetup() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// connecting to the database instance
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	// Verifing the connection
	if err = client.Ping(context.TODO(), nil); err != nil {
		log.Println("Failed to connect with database", err.Error())
	}

	fmt.Println("Connection Successful to MONGO")
	return client
}

var Client *mongo.Client = DBSetup()

func UserData(client *mongo.Client, collectionName string) *mongo.Collection {
	var userCollection *mongo.Collection = client.Database("Ecommerce").Collection(collectionName)
	return userCollection
}

func ProductData(client *mongo.Client, collectionName string) *mongo.Collection {
	var productCollection *mongo.Collection = client.Database("Ecommerce").Collection(collectionName)
	return productCollection
}
