package friendrequest

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/google/uuid"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

type friendReqCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*friendReqCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *friendReqCtrler {
	return &friendReqCtrler{argClient: argClient}
}

// TODO:
// ReqDate time need to be changed once we figure out a solution for ttime differencing issue.
func (fr *friendReqCtrler) CreateFriendReqEdge(ctx context.Context, doc *FriendRequest) error {
	var err error
	doc.Key = uuid.New().String() // Generate UUID key
	doc.Kind = kind
	_, err = fr.argClient.Coll.CreateDocument(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to create friend request edge document for requestor id %s due to %v", doc.From, err)
	}
	return nil
}

func (fr *friendReqCtrler) MkFriendReqDocId(key string) string {
	return fmt.Sprintf("%s/%s", FriendReqColl, key)
}

// Update the fields in friend request edge. Return an error if edge is not existing
func (fr *friendReqCtrler) UpdateFriendReq(ctx context.Context, key string, updateDoc map[string]interface{}) error {
	newDoc, err := convertBody(updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update friend request edge, key: %s due to %v", key, err)
	}
	_, err = fr.argClient.Coll.UpdateDocument(ctx, key, newDoc)
	if err != nil {
		return fmt.Errorf("failed to update friend request edge, key: %s due to %v", key, err)
	}
	return nil
}

func (fr *friendReqCtrler) IsFriendReqExist(ctx context.Context, query string, bindVars map[string]interface{}) (bool, error) {
	cursor, err := fr.argClient.Db.Query(ctx, query, bindVars)
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

func (fr *friendReqCtrler) RemoveFriendReqEdge(ctx context.Context, key string) error {
	_, err := fr.argClient.Coll.RemoveDocument(ctx, key)
	return err
}

func convertBody(doc map[string]interface{}) (map[string]string, error) {
	newDoc := map[string]string{}
	for key, val := range doc {
		if key == mode || key == state || key == reqDate {
			temp, ok := val.(string)
			if !ok {
				return newDoc, fmt.Errorf("invalid value name for key %s", key)
			}
			newDoc[key] = temp
			continue
		}
		return newDoc, fmt.Errorf("invalid key field %s", key)
	}
	return newDoc, nil
}
