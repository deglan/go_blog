package main

import (
	"context"
	"log"
	"social/internal/env"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoUrl = env.GetString("MONGODB_URI", "mongodb://localhost:27017")

func main() {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoUrl))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	log.Println("Database connected successfully!")

	db := client.Database("analytics")

	err = db.Collection("__init").Drop(context.Background())
	if err != nil {
		log.Fatal("Failed to ensure database exists:", err)
	}

	log.Println("✅ Database `analytics` is ready!")

	createTagsSchema(db)
}

func createTagsSchema(db *mongo.Database) {
	collection := db.Collection("tags_collection")
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "count", Value: -1},
		},
		Options: options.Index().SetUnique(false),
	}
	_, _ = collection.Indexes().CreateOne(context.Background(), indexModel)

	log.Println("✅ Created `tags_collection`")
}
