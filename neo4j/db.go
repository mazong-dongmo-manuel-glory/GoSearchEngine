package neo4j

import (
	"context"
	"fmt"

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
	fmt.Printf("Conntected ")
	db := &Database{Driver: driver}
	return db, nil

}
