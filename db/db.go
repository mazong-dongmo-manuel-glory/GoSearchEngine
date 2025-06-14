package db

import (
	"context"
	"fmt"
	"log"
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
	WordCollection     *mongo.Collection
	WordPageCollection *mongo.Collection
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
	wordPageCollection := client.Database(dbName).Collection("word_pages")
	wordPageIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "word", Value: 1},
				{Key: "page_url", Value: -1},
			},
		},
	}
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
	if _, err := wordPageCollection.Indexes().CreateMany(ctx, wordPageIndexes); err != nil {
		return nil, err
	}
	_, err = wordPageCollection.Indexes().CreateMany(ctx, wordPageIndexes)

	return &Storage{
		Client:             client,
		PageCollection:     pageCollection,
		WordPageCollection: wordPageCollection,
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
			documents[i] = bson.D{{Key: "url", Value: url}}
		}
		_, err := s.UrlQueueCollection.InsertMany(ctx, documents, options.InsertMany().SetOrdered(false))
		if err != nil {
			panic(err)
			return
		}
	default:

	}

}
func (s *Storage) StoreMany(interfaces []interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fmt.Println(interfaces[0])

	_, err := s.WordPageCollection.InsertMany(ctx, interfaces, options.InsertMany().SetOrdered(false))
	if err != nil {
		panic(err)
	}
}

func GetPages(storage *Storage, limit int64) []*Page {
	var pages []*Page

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
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
	collection := s.Client.Database("search_engine").Collection("word_pages") // En supposant que "word_pages" est le nom de votre collection

	// Requête pour les WordPages où le champ 'word' est dans le slice 'words' fourni.
	filter := bson.M{"word": bson.M{"$in": words}}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "tfidf", Value: -1}}) // Trier par TF-IDF descendant
	findOptions.SetLimit(limit)                            // Limiter le nombre de résultats par mot si nécessaire

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

// CountPages retourne le nombre total de pages
func CountPages(storage *Storage) int {
	// Implémentation dépendante de votre base de données
	// Exemple pour SQL :
	/*
	   var count int
	   err := storage.db.QueryRow("SELECT COUNT(*) FROM pages").Scan(&count)
	   if err != nil {
	       return 0
	   }
	   return count
	*/
	return 0 // À implémenter selon votre DB
}

// GetPagesWithOffset récupère les pages avec LIMIT et OFFSET
func GetPagesWithOffset(storage *Storage, limit, offset int) []*Page {
	// Implémentation dépendante de votre base de données
	// Exemple pour SQL :
	/*
	   rows, err := storage.db.Query("SELECT url, content, page_rank FROM pages LIMIT ? OFFSET ?", limit, offset)
	   if err != nil {
	       return nil
	   }
	   defer rows.Close()

	   var pages []*Page
	   for rows.Next() {
	       page := &Page{}
	       err := rows.Scan(&page.Url, &page.Content, &page.PageRank)
	       if err != nil {
	           continue
	       }
	       pages = append(pages, page)
	   }
	   return pages
	*/
	return nil // À implémenter selon votre DB
}
