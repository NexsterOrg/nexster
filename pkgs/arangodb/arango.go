package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

type Config struct {
	Hostname, Database, Username, Password string
	Port                                   int
}

type Client struct {
	Db   driver.Database
	Coll driver.CollectionDocuments
}

func newClient(cfg *Config) driver.Client {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{fmt.Sprintf("http://%s:%d", cfg.Hostname, cfg.Port)},
	})
	if err != nil {
		panic(err)
	}
	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(cfg.Username, cfg.Password),
	})
	if err != nil {
		panic(err)
	}
	return client
}

// Create new client to work with db
func NewDbClient(ctx context.Context, cfg *Config) *Client {
	client := newClient(cfg)
	db, err := client.Database(ctx, cfg.Database)
	if err != nil {
		panic(err)
	}
	return &Client{Db: db}
}

// Create new client to work with specific collection
func NewCollClient(ctx context.Context, cfg *Config, collection string) *Client {
	client := newClient(cfg)
	db, err := client.Database(ctx, cfg.Database)
	if err != nil {
		panic(err)
	}
	coll, err := db.Collection(ctx, collection)
	if err != nil {
		panic(err)
	}
	return &Client{Db: db, Coll: coll}
}
