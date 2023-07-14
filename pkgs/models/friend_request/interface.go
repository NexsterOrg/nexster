package friendrequest

import "context"

const FriendReqColl string = "friendRequest"

const (
	kind    string = "friend_request"
	mode    string = "mode"
	state   string = "state"
	reqDate string = "req_state"
)

// This is Friend Request data model
type Interface interface {
	CreateFriendReqEdge(ctx context.Context, doc *FriendRequest) (string, error)
	UpdateFriendReq(ctx context.Context, key string, updateDoc map[string]interface{}) error
	MkFriendReqDocId(key string) string
	IsFriendReqExist(ctx context.Context, query string, bindVars map[string]interface{}) (bool, error)
	RemoveFriendReqEdge(ctx context.Context, key string) error
}

type FriendRequest struct {
	Key     string `json:"_key"`
	From    string `json:"_from"`
	To      string `json:"_to"`
	Kind    string `json:"kind"`
	Mode    string `json:"mode"`
	State   string `json:"state"`
	ReqDate string `json:"req_date"`
}
