package friend

import "context"

const kind string = "friend" // kind of edge
const FriendColl string = "friends"

// friend types
const (
	FriendType           string = "friend"
	NotFriendType        string = "not_friend"
	PendingReqstorType   string = "pending#requestor"
	PendingRecipientType string = "pending#recipient"
)

type Interface interface {
	CreateFriendEdge(ctx context.Context, doc *Friend) error
	MkFriendDocId(key string) string
	RemoveFriendEdge(ctx context.Context, key string) error
	IsFriendEdgeExist(ctx context.Context, user1, user2 string) (bool, error)
	CountFriends(ctx context.Context, query string, bindVars map[string]interface{}) (int, error)
	GetShortestDistance(ctx context.Context, startNodeKey, endNodeKey string) (int, error)
}

// document format for `friend` edge
type Friend struct {
	Key           string `json:"_key"`
	From          string `json:"_from"`
	To            string `json:"_to"`
	Kind          string `json:"kind"`
	StartedAt     string `json:"started_at"`
	OtherFriendId string `json:"other_friend_id"`
}
