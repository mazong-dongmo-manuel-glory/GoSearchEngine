package neo4j

import (
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"search_egine/models"
)

type ConfigStruct struct {
	Host     string
	Username string
	Password string
	Ctx      context.Context
}

var Config ConfigStruct = ConfigStruct{
	Host:     "neo4j://localhost:7687/neo4j",
	Username: "neo4j",
	Password: "M@zong2003",
	Ctx:      context.Background(),
}

type Database struct {
	Driver neo4j.DriverWithContext
}

func NewStorage() (*Database, error) {
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(Config.Host, neo4j.BasicAuth(Config.Username, Config.Password, ""))
	if err != nil {
		return nil, err
	}

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Conntected\n")
	db := &Database{Driver: driver}
	return db, nil

}

func (db *Database) Save(page models.Page) error {
	_, err := neo4j.ExecuteQuery(Config.Ctx, db.Driver, `
MERGE (main:Page {url: $url})
SET 
    main.pagerank = $pagerank,
    main.content = $content,
    main.urls = $urls

RETURN NULL

`, map[string]any{
		"url":      page.Url,
		"content":  page.Content,
		"urls":     page.Urls,
		"pagerank": 0,
	}, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase("neo4j"))

	if err != nil {
		panic(err)
	}
	return err
}
