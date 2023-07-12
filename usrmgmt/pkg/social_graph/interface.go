package socialgraph

import "context"

type Interface interface {
	CreateFriendReq(ctx context.Context, reqstorKey, friendKey, mode, state, reqDate string) error
	RemoveFriendRequest(ctx context.Context, key string) error
}
