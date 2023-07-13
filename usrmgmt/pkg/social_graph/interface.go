package socialgraph

import "context"

type Interface interface {
	CreateFriendReq(ctx context.Context, reqstorKey, friendKey, mode, state, reqDate string) (map[string]string, error)
	RemoveFriendRequest(ctx context.Context, key string) error
	CreateFriend(ctx context.Context, friendReqKey, user1, user2, acceptedAt string) (map[string]string, error)
}
