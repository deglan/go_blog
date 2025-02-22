package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoStorage struct {
	Tags interface {
		UpdateTagsUsage(context.Context, []string) error
		GetTrendingTags(context.Context, int) ([]Tag, error)
	}
}

func NewMongoStorage(db *mongo.Database) MongoStorage {
	return MongoStorage{
		Tags: &MongoTagStore{db: db},
	}
}
