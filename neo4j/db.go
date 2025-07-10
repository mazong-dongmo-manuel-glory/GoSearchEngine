package neo4j

import (
	"context"
	"fmt"
	"search_egine/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
	defer driver.Close(Config.Ctx)

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Conntected\n")
	db := &Database{Driver: driver}
	return db, nil

}

func (db *Database) Save(node interface{}) {
	switch n := node.(type) {
	case models.Page:

		res, err := neo4j.ExecuteQuery(Config.Ctx, db.Driver, `MERGE (p:Page {url : $url, title : $title, pagerank : $pagerank) RETURN p`, map[string]any{
			"url":      n.Url,
			"title":    n.Title,
			"pagerank": n.PageRank,
		}, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase("neo4j"))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v\n", res)
	}
}
