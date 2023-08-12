package socialgraph

import (
	"context"

	usr "github.com/NamalSanjaya/nexster/pkgs/models/user"
)

type Interface interface {
	CreateFriendReq(ctx context.Context, reqstorKey, friendKey, mode, state, reqDate string) (map[string]string, error)
	RemoveFriendRequest(ctx context.Context, key string) error
	CreateFriend(ctx context.Context, friendReqKey, user1, user2, acceptedAt string) (map[string]string, error)
	RemoveFriend(ctx context.Context, key1, key2 string) error
	ListFriends(ctx context.Context, userId string, offset, count int) ([]*map[string]string, error)
	CountFriends(ctx context.Context, userId string) (int, error)
	GetRole(authUserKey, userKey string) usr.UserRole
}
