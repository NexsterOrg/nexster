package event

import (
	"context"
	"fmt"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	utm "github.com/NamalSanjaya/nexster/pkgs/utill/time"
)

type eventCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*eventCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *eventCtrler {
	return &eventCtrler{argClient: argClient}
}

func (ev *eventCtrler) MkDocumentId(key string) string {
	return fmt.Sprintf("%s/%s", EventColl, key)
}

func (ev *eventCtrler) CreateDocument(ctx context.Context, doc *Event) (string, error) {
	// setting default parameters
	doc.Kind = kind
	doc.Visibility = visibility
	doc.CreatedAt = utm.CurrentUTCTimeTillMinutes()

	meta, err := ev.argClient.Coll.CreateDocument(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to create friend request edge document for requestor id due to %v", err)
	}
	return meta.Key, nil
}
