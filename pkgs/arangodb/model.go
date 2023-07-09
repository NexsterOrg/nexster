package arangodb

import (
	"context"
)

type Interface interface {
	ListMedia(ctx context.Context, query string, bindVars map[string]interface{}) ([]*Media, error)
}

type Media struct {
	Link        string `json:"link"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreateDate  string `json:"created_date"`
	Size        int    `json:"size"`
}
