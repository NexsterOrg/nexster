package media

import (
	"context"
)

const MediaColl string = "media" // Name of collection

type Interface interface {
	MkMediaDocId(key string) string
	ListMedia(ctx context.Context, query string, bindVars map[string]interface{}) ([]*Media, error)
	ListMediaWithOwner(ctx context.Context, query string, bindVars map[string]interface{}) ([]*MediaWithOwner, error)
	ListMediaWithCustomFields(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]string, error)
	Get(ctx context.Context, key string) (*Media, error)
}

type Media struct {
	Key         string `json:"_key"`
	Link        string `json:"link"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreateDate  string `json:"created_date"`
	Size        int    `json:"size"`
	Visibility  string `json:"visibility"`
}

type MediaWithOwner struct {
	Media   Media  `json:"media"`
	OwnerId string `json:"owner_id"`
}
