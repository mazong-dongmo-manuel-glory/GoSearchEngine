package neo4jtest

import (
	"search_egine/models"
	"search_egine/neo4j"
	"testing"
)

func TestNewStorage(t *testing.T) {
	db, err := neo4j.NewStorage()
	if err != nil {
		t.Fatalf("Error on db %v", err)
	}
	t.Log(db)
}

func TestDatabaseSave(t *testing.T) {
	db, err := neo4j.NewStorage()
	if err != nil {
		panic(err)
	}
	p := models.Page{
		Url:      "https://www.test.com",
		Title:    "test",
		PageRank: 0.5556565778945,
	}
	db.Save(p)
}
