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

type Client struct {
	db driver.Database
}

var _ Interface = (*Client)(nil)

// create new client with arango db
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
	return &Client{db: db}
}

func (cl *Client) ListMedia(ctx context.Context, query string, bindVars map[string]interface{}) ([]*Media, error) {
	var posts []*Media
	cursor, err := cl.db.Query(ctx, query, bindVars)
	if err != nil {
		return posts, err
	}
	defer cursor.Close()

	for {
		var post Media
		_, err := cursor.ReadDocument(ctx, &post)
		if driver.IsNoMoreDocuments(err) {
			return posts, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		posts = append(posts, &post)
	}
}

func (cl *Client) ListUsers(ctx context.Context, query string, bindVars map[string]interface{}) ([]*User, error) {
	var users []*User
	cursor, err := cl.db.Query(ctx, query, bindVars)
	if err != nil {
		return users, err
	}
	defer cursor.Close()

	for {
		var user User
		_, err := cursor.ReadDocument(ctx, &user)
		if driver.IsNoMoreDocuments(err) {
			return users, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		users = append(users, &user)
	}
}
