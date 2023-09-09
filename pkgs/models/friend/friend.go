package friend

import (
	"context"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

const friendEdgeExistQuery string = `FOR v,e,p IN 1..1 ANY
	@fromUser friends
	OPTIONS { uniqueVertices: "path" }
	FILTER e._to == @toUser
	return e._key`

const getShortestDistanceQry string = `FOR v IN OUTBOUND SHORTEST_PATH 
	@startNode TO @endNode friends RETURN "1"`

const listFriendsQry string = `FOR v,e IN 1..1 OUTBOUND
	@startNode friends
	SORT e.started_at DESC
	LIMIT @offset, @count
	RETURN {
		"user_id": v._key,
		"name": v.username,
		"from_friend_id": e._key,
		"to_friend_id": e.other_friend_id,
		"image_url" : v.image_url,
		"faculty" : v.faculty,
		"field" : v.field,
		"batch" : v.batch,
		"index_no" : v.index_no
	}`

const removeFriendshipQry string = `FOR frnd IN friends
	FILTER (frnd._from == @nodeId1 && frnd._to == @nodeId2)
	REMOVE frnd IN friends
	RETURN frnd._key`

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

func (frn *friendCtrler) RemoveFriendship(ctx context.Context, userId1, userId2 string) (map[string]string, error) {
	result := map[string]string{"id1": "", "id2": ""}
	// remove one direction friend link
	rm1Ids, err := frn.listStrings(ctx, removeFriendshipQry, map[string]interface{}{
		"nodeId1": userId1,
		"nodeId2": userId2,
	})
	if err != nil {
		return result, err
	}
	ln1 := len(rm1Ids)
	if ln1 == 0 {
		return result, nil
	}
	if ln1 > 1 {
		return result, fmt.Errorf("more than once edage from: %s to %s", userId1, userId2)
	}
	result["id1"] = *(rm1Ids[0])
	// remove other friend link
	rm2Ids, err := frn.listStrings(ctx, removeFriendshipQry, map[string]interface{}{
		"nodeId1": userId2,
		"nodeId2": userId1,
	})
	if err != nil {
		return result, err
	}
	ln2 := len(rm2Ids)
	if ln2 == 0 {
		return result, nil
	}
	if ln2 > 1 {
		return result, fmt.Errorf("more than once edage from: %s to %s", userId2, userId1)
	}
	result["id2"] = *(rm2Ids[0])
	return result, nil
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

// count friend edges
func (frn *friendCtrler) CountFriends(ctx context.Context, query string, bindVars map[string]interface{}) (int, error) {
	cursor, err := frn.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return 0, err
	}
	defer cursor.Close()

	for {
		var count int
		_, err := cursor.ReadDocument(ctx, &count)
		if driver.IsNoMoreDocuments(err) {
			return 0, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		return count, nil
	}
}

func (frn *friendCtrler) GetShortestDistance(ctx context.Context, startNodeKey, endNodeKey string) (int, error) {
	path, err := frn.listStrings(ctx, getShortestDistanceQry, map[string]interface{}{
		"startNode": startNodeKey,
		"endNode":   endNodeKey,
	})
	return len(path), err
}

func (frn *friendCtrler) ListFriends(ctx context.Context, userId string, offset, count int) ([]*map[string]string, error) {
	return frn.listJsonWithStringFields(ctx, listFriendsQry, map[string]interface{}{
		"startNode": userId,
		"offset":    offset,
		"count":     count,
	})
}

// Return list of strings. eg: ["elem1", "elem2", "elem3" ]
func (frn *friendCtrler) listStrings(ctx context.Context, query string, bindVar map[string]interface{}) ([]*string, error) {
	var results []*string
	cursor, err := frn.argClient.Db.Query(ctx, query, bindVar)
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

// list json whose fields are strings. eg: [{}, {}, {}]. each json has key-value pair, value being string.
func (frn *friendCtrler) listJsonWithStringFields(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]string, error) {
	results := []*map[string]string{}
	cursor, err := frn.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result map[string]string
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
