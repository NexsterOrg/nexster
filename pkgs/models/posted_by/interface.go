package postedby

import (
	"context"
)

const PostedByColl string = "postedBy" // name of the postedBy edge in arango db

const (
	kind      string = "postedBy"
	TypeEvent string = "event"
)

type Interface interface {
	MkDocumentId(key string) string
	CreateDocument(ctx context.Context, from, to, type1 string) (string, error)
}

type PostedBy struct {
	Key  string `json:"_key"`
	From string `json:"_from"`
	To   string `json:"_to"`
	Kind string `json:"kind"`
	Type string `json:"type"`
}
