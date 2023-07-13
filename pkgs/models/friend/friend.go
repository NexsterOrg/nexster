package friend

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

type friendCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*friendCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *friendCtrler {
	return &friendCtrler{argClient: argClient}
}

// TODO:
// ReqDate time need to be changed once we figure out a solution for ttime differencing issue.
func (frn *friendCtrler) CreateFriendEdge(ctx context.Context, doc *Friend) (string, error) {
	doc.Key = uuid.New().String() // Generate UUID key
	doc.Kind = kind
	_, err := frn.argClient.Coll.CreateDocument(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to create friend edge document for fromKey: %s, toKey: %s id due to %v", doc.From, doc.To, err)
	}
	return doc.Key, nil
}

func (frn *friendCtrler) MkFriendDocId(key string) string {
	return fmt.Sprintf("%s/%s", FriendColl, key)
}

func (frn *friendCtrler) RemoveFriendEdge(ctx context.Context, key string) error {
	_, err := frn.argClient.Coll.RemoveDocument(ctx, key)
	return err
}
