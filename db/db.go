package db

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Document interface {
	Save()
}
type Storage struct {
	DBName             string
	Client             *mongo.Client
	PageCollection     *mongo.Collection
	UrlQueueCollection *mongo.Collection
}

func NewStorage(dbName string) (*Storage, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}
	pageCollection := client.Database(dbName).Collection("pages")
	urlQueueCollection := client.Database(dbName).Collection("urls")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "url", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err = pageCollection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return nil, err
	}
	return &Storage{
		Client:             client,
		PageCollection:     pageCollection,
		UrlQueueCollection: urlQueueCollection,
	}, nil

}

func (s *Storage) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.Client.Disconnect(ctx)
}

func (s *Storage) Store(d interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	switch d.(type) {
	case *Page:
		_, err := s.PageCollection.InsertOne(ctx, d.(*Page))
		if err != nil {
			return
		}
	case []string:
		documents := make([]interface{}, len(d.([]string)))
		for i, url := range d.([]string) {
			b := UrlQueuElement{Url: url}
			documents[i] = b
		}
		_, err := s.UrlQueueCollection.InsertMany(ctx, documents, options.InsertMany().SetOrdered(false))
		if err != nil {
			panic(err)
			return
		}

	}

}
