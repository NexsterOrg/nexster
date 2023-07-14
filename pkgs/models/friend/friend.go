package friend

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

const friendEdgeExistQuery string = `FOR v,e,p IN 1..1 ANY
	@fromUser friends
	OPTIONS { uniqueVertices: "path" }
	FILTER e._to == @toUser
	return e._key`

type friendCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*friendCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *friendCtrler {
	return &friendCtrler{argClient: argClient}
}

// TODO:
// ReqDate time need to be changed once we figure out a solution for ttime differencing issue.
func (frn *friendCtrler) CreateFriendEdge(ctx context.Context, doc *Friend) error {
	doc.Kind = kind
	_, err := frn.argClient.Coll.CreateDocument(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to create friend edge document for fromKey: %s, toKey: %s id due to %v", doc.From, doc.To, err)
	}
	return nil
}

func (frn *friendCtrler) MkFriendDocId(key string) string {
	return fmt.Sprintf("%s/%s", FriendColl, key)
}

func (frn *friendCtrler) RemoveFriendEdge(ctx context.Context, key string) error {
	_, err := frn.argClient.Coll.RemoveDocument(ctx, key)
	return err
}

func (frn *friendCtrler) IsFriendEdgeExist(ctx context.Context, user1, user2 string) (bool, error) {
	cursor, err := frn.argClient.Db.Query(ctx, friendEdgeExistQuery, map[string]interface{}{
		"fromUser": user1, "toUser": user2,
	})
	if err != nil {
		return false, err
	}
	defer cursor.Close()

	for {
		var key string
		_, err := cursor.ReadDocument(ctx, &key)
		if driver.IsNoMoreDocuments(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		return true, nil
	}
}
