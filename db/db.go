package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Document interface {
	Save()
}
type Storage struct {
	Client     *mongo.Client
	collection *mongo.Collection
}

func NewStorage() (*Storage, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}
	collection := client.Database("search_engine").Collection("pages")

	//creation des index
	/*
		indexes := []mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "url", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{{Key: "domain", Value: 1}},
			},
		}
		_, err = collection.Indexes().CreateMany(ctx, indexes)
		if err != nil {
			return nil, err
		}*/
	return &Storage{
		Client:     client,
		collection: collection,
	}, nil

}

func (s *Storage) Store(d Document) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := s.collection.InsertOne(ctx, d)
	if err != nil {
		panic(err)
	}
}
