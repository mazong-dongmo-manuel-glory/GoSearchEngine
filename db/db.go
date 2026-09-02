package db

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConfigStruct struct {
	Host     string
	Port     int
	Database string
}

var Config ConfigStruct = ConfigStruct{
	Host:     "localhost",
	Port:     27017,
	Database: "search_engine",
}

type Document interface {
	Save()
}
type Storage struct {
	DBName             string
	Client             *mongo.Client
	PageCollection     *mongo.Collection
	UrlQueueCollection *mongo.Collection
	WordCollection     *mongo.Collection
	WordPageCollection *mongo.Collection
}

// ── Singleton global ────────────────────────────────────────────────
// Le driver MongoDB gère déjà un pool de connexions en interne.
// On ne crée qu'UN seul client pour tout le programme.
var (
	globalStorage *Storage
	storageOnce   sync.Once
	storageErr    error
)

func GetGlobalStorage() (*Storage, error) {
	storageOnce.Do(func() {
		globalStorage, storageErr = newStorage()
	})
	return globalStorage, storageErr
}

func newStorage() (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+Config.Host+":"+fmt.Sprintf("%d", Config.Port)))
	if err != nil {
		return nil, err
	}
	pageCollection := client.Database(Config.Database).Collection("pages")
	urlQueueCollection := client.Database(Config.Database).Collection("urls")
	wordPageCollection := client.Database(Config.Database).Collection("word_pages")

	wordPageIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "word", Value: 1},
				{Key: "page_url", Value: -1},
			},
		},
	}
	pageIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "url", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err = pageCollection.Indexes().CreateMany(ctx, pageIndexes)
	if err != nil {
		return nil, err
	}
	// Un seul appel — l'ancien code le faisait 2 fois
	if _, err := wordPageCollection.Indexes().CreateMany(ctx, wordPageIndexes); err != nil {
		return nil, err
	}

	return &Storage{
		Client:             client,
		PageCollection:     pageCollection,
		WordPageCollection: wordPageCollection,
		UrlQueueCollection: urlQueueCollection,
	}, nil
}

// NewStorage conservé pour compatibilité, mais redirige vers le singleton
func NewStorage() (*Storage, error) {
	return GetGlobalStorage()
}

// Close ne fait rien pour le singleton — le client vit toute la durée du programme
func (s *Storage) Close() {}

func (s *Storage) Store(d interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	switch v := d.(type) {
	case *Page:
		_, err := s.PageCollection.InsertOne(ctx, v)
		if err != nil {
			return
		}
	case []string:
		documents := make([]interface{}, len(v))
		for i, url := range v {
			documents[i] = bson.D{{Key: "url", Value: url}}
		}
		_, err := s.UrlQueueCollection.InsertMany(ctx, documents, options.InsertMany().SetOrdered(false))
		if err != nil {
			log.Printf("Erreur InsertMany urls: %v", err)
			return
		}
	}
}

func (s *Storage) StoreMany(interfaces []interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err := s.WordPageCollection.InsertMany(ctx, interfaces, options.InsertMany().SetOrdered(false))
	if err != nil {
		log.Printf("Erreur StoreMany: %v", err)
	}
}

func GetPages(storage *Storage, limit int64) []*Page {
	var pages []*Page

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cursor, err := storage.PageCollection.Find(ctx, bson.D{}, options.Find().SetLimit(limit))
	if err != nil {
		log.Printf("Error finding pages: %v", err)
		return nil
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var page Page
		if err := cursor.Decode(&page); err != nil {
			log.Printf("Error decoding page: %v", err)
			continue
		}
		pages = append(pages, &page)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
	}

	return pages
}

func (s *Storage) GetWordPagesByWords(words []string, limit int64) ([]WordPage, error) {
	collection := s.WordPageCollection

	filter := bson.M{"word": bson.M{"$in": words}}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "tfidf", Value: -1}})
	findOptions.SetLimit(limit)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find word pages: %v", err)
	}
	defer cursor.Close(ctx)

	var results []WordPage
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode word pages: %v", err)
	}
	return results, nil
}

func (s *Storage) UpdatePageRank(url string, rank float64) error {
	collection := s.PageCollection
	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"url": url},
		bson.M{"$set": bson.M{"pagerank": rank}},
	)
	return err
}

// BulkUpdatePageRanks met à jour tous les PageRanks en un seul appel réseau
// au lieu de N appels UpdateOne séparés.
func (s *Storage) BulkUpdatePageRanks(ranks map[string]float64) error {
	if len(ranks) == 0 {
		return nil
	}

	models := make([]mongo.WriteModel, 0, len(ranks))
	for url, rank := range ranks {
		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"url": url}).
			SetUpdate(bson.M{"$set": bson.M{"pagerank": rank}})
		models = append(models, model)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, err := s.PageCollection.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	return err
}
