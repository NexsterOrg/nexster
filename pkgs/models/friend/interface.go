package friend

import "context"

const kind string = "friend" // kind of edge
const FriendColl string = "friends"

type Interface interface {
	CreateFriendEdge(ctx context.Context, doc *Friend) (string, error)
	MkFriendDocId(key string) string
	RemoveFriendEdge(ctx context.Context, key string) error
	IsFriendEdgeExist(ctx context.Context, user1, user2 string) (bool, error)
}

// document format for `friend` edge
type Friend struct {
	Key       string `json:"_key"`
	From      string `json:"_from"`
	To        string `json:"_to"`
	Kind      string `json:"kind"`
	StartedAt string `json:"started_at"`
}
