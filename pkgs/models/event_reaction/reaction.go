package eventreaction

import (
	"context"
	"fmt"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	"github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
)

type eventReactionCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*eventReactionCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *eventReactionCtrler {
	return &eventReactionCtrler{argClient: argClient}
}

func MkDocumentId(key string) string {
	return fmt.Sprintf("%s/%s", EventReactionColl, key)
}

func (erc *eventReactionCtrler) Get(ctx context.Context, key string) (*EventReaction, error) {
	reaction := &EventReaction{}
	_, err := erc.argClient.Coll.ReadDocument(ctx, key, reaction)
	return reaction, err
}

func (erc *eventReactionCtrler) Create(ctx context.Context, data *EventReaction) (string, error) {
	key := uuid.GenUUID4()
	data.Key = key
	_, err := erc.argClient.Coll.CreateDocument(ctx, data)
	if err != nil {
		return "", fmt.Errorf("failed to create event reaction link: %v", err)
	}
	return key, nil
}
