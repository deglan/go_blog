package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Tag struct {
	ID       string    `bson:"_id"`
	Count    int       `bson:"count"`
	LastUsed time.Time `bson:"last_used"`
}

type MongoTagStore struct {
	db *mongo.Database
}

func (s *MongoTagStore) UpdateTagsUsage(ctx context.Context, tags []string) error {
	collection := s.db.Collection("tags_collection")

	for _, tag := range tags {
		filter := bson.M{"_id": tag} // Wyszukujemy po nazwie tagu
		update := bson.M{
			"$inc": bson.M{"count": 1},              // Zwiększamy licznik
			"$set": bson.M{"last_used": time.Now()}, // Aktualizujemy datę
		}

		opts := options.Update().SetUpsert(true).SetBypassDocumentValidation(false)

		_, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MongoTagStore) GetTrendingTags(ctx context.Context, limit int) ([]Tag, error) {
	collection := s.db.Collection("tags_collection")

	opts := options.Find().SetSort(bson.M{"count": -1}).SetLimit(int64(limit))
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tags []Tag
	if err := cursor.All(ctx, &tags); err != nil {
		return nil, err
	}

	return tags, nil
}
