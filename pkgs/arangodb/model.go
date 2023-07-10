package arangodb

import (
	"context"
)

const (
	UsersDoc string = "users"
)

type Interface interface {
	ListMedia(ctx context.Context, query string, bindVars map[string]interface{}) ([]*Media, error)
	ListUsers(ctx context.Context, query string, bindVars map[string]interface{}) ([]*User, error)
	CreateDocId(doc, key string) string
}

type Media struct {
	Link        string `json:"link"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreateDate  string `json:"created_date"`
	Size        int    `json:"size"`
}

type User struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	ImageUrl string `json:"image_url"`
}
