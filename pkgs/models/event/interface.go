package event

import (
	"context"
)

const EventColl string = "events" // name of the event collection in arango db

// default parameters
const (
	kind       string = "event"
	visibility string = "public"
)

type Interface interface {
	MkDocumentId(key string) string
	CreateDocument(ctx context.Context, doc *Event) (string, error)
	ListUpcomingsByDate(ctx context.Context, offset, count int) ([]*map[string]string, error)
	Get(ctx context.Context, key string) (*Event, error)
	Delete(ctx context.Context, eventKey string) (*map[string]interface{}, error)
}

type Event struct {
	Key         string `json:"_key"`
	Link        string `json:"link"`
	Kind        string `json:"kind"`
	ImgType     string `json:"imgType"`
	Visibility  string `json:"visibility"`
	Title       string `json:"title"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Venue       string `json:"venue"`
	Mode        string `json:"mode"`
	EventLink   string `json:"eventLink"`
	CreatedAt   string `json:"createdAt"`
}
