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
	Db driver.Database
}

// Create new client with arango db
func NewDbClient(ctx context.Context, cfg *Config) *Client {
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

	db, err := client.Database(ctx, cfg.Database)
	if err != nil {
		panic(err)
	}
	return &Client{Db: db}
}
