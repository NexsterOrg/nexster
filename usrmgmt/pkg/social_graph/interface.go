package socialgraph

import (
	"context"

	usr "github.com/NamalSanjaya/nexster/pkgs/models/user"
)

type Interface interface {
	CreateFriendReq(ctx context.Context, reqstorKey, friendKey, mode, state, reqDate string) (map[string]string, error)
	RemoveFriendRequest(ctx context.Context, friendkey, user1Key, user2Key string) error
	CreateFriend(ctx context.Context, friendReqKey, user1, user2, acceptedAt string) (map[string]string, error)
	RemoveFriend(ctx context.Context, key1, key2 string) error
	RemoveFriendV2(ctx context.Context, userKey1, userKey2 string) (map[string]string, error)
	ListFriends(ctx context.Context, userId string, offset, count int) ([]*map[string]string, error)
	CountFriends(ctx context.Context, userId string) (int, error)
	GetRole(authUserKey, userKey string) usr.UserRole
	GetProfileInfo(ctx context.Context, userKey string) (map[string]string, error)
	CountFriendsV2(ctx context.Context, userId string) (int, error)
	GetUserKeyByIndexNo(ctx context.Context, indexNo string) (string, error)
	ListFriendReqs(ctx context.Context, userKey string, offset, count int) ([]*map[string]string, error)
	GetAllFriendReqsCount(ctx context.Context, userKey string) (int, error)
	UpdateUser(ctx context.Context, userId string, data map[string]interface{}) error
	DeleteUser(ctx context.Context, userId string) error
}
