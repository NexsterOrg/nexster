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
	CreateForGivenKey(ctx context.Context, data *Media) (string, error)
	ListStrings(ctx context.Context, query string, bindVars map[string]interface{}) ([]string, error)
	DeleteDocument(ctx context.Context, mediaKey string) (map[string]interface{}, error)
}

// TODO: Add kind = "media" if needed.
type Media struct {
	Key         string `json:"_key"`
	Link        string `json:"link"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreateDate  string `json:"created_date"`
	Size        int    `json:"size"` // NOTE: size variable is not currently used. Therefore, we setting its value to zero.
	Visibility  string `json:"visibility"`
}

type MediaWithOwner struct {
	Media   Media  `json:"media"`
	OwnerId string `json:"owner_id"`
}
