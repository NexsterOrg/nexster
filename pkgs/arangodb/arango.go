package arangodb

import (
	"context"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

type Config struct {
	Hostname, Database, Username, Password string
	Port                                   int
}

type dbClient struct {
	db driver.Database
}

// TODO:
// Need two seperate clients for Db and collection
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

// Create new client to work with complete Db
func NewDbClient(ctx context.Context, cfg *Config) *dbClient {
	client := newClient(cfg)
	db, err := client.Database(ctx, cfg.Database)
	if err != nil {
		panic(err)
	}
	return &dbClient{db: db}
}

// results format [ {}, {}, {} ]
func (d *dbClient) ListJsonAnyValue(ctx context.Context, query string, bindVar map[string]interface{}) ([]*map[string]interface{}, error) {
	results := []*map[string]interface{}{}
	cursor, err := d.db.Query(ctx, query, bindVar)
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &result)
		if driver.IsNoMoreDocuments(err) {
			return results, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		results = append(results, &result)
	}
}

// results format [ "elem1", "elem2", "elem3" ]
func (d *dbClient) ListStrings(ctx context.Context, query string, bindVar map[string]interface{}) ([]string, error) {
	results := []string{}
	cursor, err := d.db.Query(ctx, query, bindVar)
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result string
		_, err := cursor.ReadDocument(ctx, &result)
		if driver.IsNoMoreDocuments(err) {
			return results, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		results = append(results, result)
	}
}
