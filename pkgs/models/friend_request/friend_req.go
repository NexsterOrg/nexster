package friendrequest

import (
	"context"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"
	"github.com/google/uuid"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

const getFriendReqKeyQry string = `FOR v,e IN 1..1 OUTBOUND
	@reqstorNode friendRequest
	FILTER v._id == @friendNode
	return e._key`

type friendReqCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*friendReqCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *friendReqCtrler {
	return &friendReqCtrler{argClient: argClient}
}

// Return list of strings (eg: ["element1", "anotherElem2", "thirdElem3"] ). This will use when Query return value is list of strings.
func (fr *friendReqCtrler) ListStrings(ctx context.Context, query string, bindVars map[string]interface{}) ([]int, error) {
	results := []int{}
	cursor, err := fr.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result int
		_, err := cursor.ReadDocument(ctx, &result)
		if driver.IsNoMoreDocuments(err) {
			return results, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		results = append(results, result)
	}
}

// eg: [ {}, {}, {} ]. Each json has string value.
func (fr *friendReqCtrler) ListStringValueJson(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]string, error) {
	results := []*map[string]string{}
	cursor, err := fr.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		result := map[string]string{}
		_, err := cursor.ReadDocument(ctx, &result)
		if driver.IsNoMoreDocuments(err) {
			return results, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		results = append(results, &result)
	}
}

// TODO:
// 1. ReqDate time need to be changed once we figure out a solution for ttime differencing issue.
// 2. Friend request edge should only be uni directional(Fix this).
func (fr *friendReqCtrler) CreateFriendReqEdge(ctx context.Context, doc *FriendRequest) (string, error) {
	var err error
	doc.Key = uuid.New().String() // Generate UUID key
	doc.Kind = kind
	_, err = fr.argClient.Coll.CreateDocument(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to create friend request edge document for requestor id %s due to %v", doc.From, err)
	}
	return doc.Key, nil
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

// TODO: Refactor the method
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

func (fr *friendReqCtrler) GetFriendReqKey(ctx context.Context, reqstorId, friendId string) (string, error) {
	fReqKeys, err := fr.listStrings(ctx, getFriendReqKeyQry, map[string]interface{}{
		"reqstorNode": reqstorId,
		"friendNode":  friendId,
	})
	if err != nil {
		return "", err
	}
	ln := len(fReqKeys)
	if ln == 0 {
		return "", nil
	}
	if ln > 1 {
		return "", fmt.Errorf("more than one friend_req keys found")
	}
	return *(fReqKeys[0]), nil
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

func (fr *friendReqCtrler) listStrings(ctx context.Context, query string, bindVar map[string]interface{}) ([]*string, error) {
	var results []*string
	cursor, err := fr.argClient.Db.Query(ctx, query, bindVar)
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result string
		_, err := cursor.ReadDocument(ctx, &result)
		if driver.IsNoMoreDocuments(err) {
			return results, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		results = append(results, &result)
	}
}
