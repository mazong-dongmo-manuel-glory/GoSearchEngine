package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27018"))
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
func (s *Storage) StoreQueue(urls []string) {

	docs := make([]interface{}, len(urls))
	for i, url := range urls {
		docs[i] = UrlQueuElement{Url: url}

	}
	_, err := s.UrlQueueCollection.InsertMany(context.Background(), docs)
	if err != nil {
		panic(err)
		return
	}
}
func (s *Storage) GetQueue(size int64) []string {
	ctx := context.Background()

	// Options pour trier par _id et limiter le nombre de résultats
	findOptions := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetLimit(size)

	// Trouver les documents avec la limite spécifiée
	cursor, err := s.UrlQueueCollection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	var results []UrlQueuElement
	if err := cursor.All(ctx, &results); err != nil {
		return nil
	}

	// Extraire les URLs
	urls := make([]string, len(results))
	for i, doc := range results {
		urls[i] = doc.Url
	}

	// Supprimer les documents récupérés
	if len(urls) > 0 {
		_, err := s.UrlQueueCollection.DeleteMany(ctx, bson.D{
			{Key: "url", Value: bson.D{{Key: "$in", Value: urls}}},
		})
		if err != nil {
			return nil
		}
	}

	return urls
}
