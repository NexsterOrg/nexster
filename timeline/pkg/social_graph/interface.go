package social_graph

import (
	"context"

	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
)

type Interface interface {
	ListRecentPosts(ctx context.Context, userId, lastPostTimestamp, visibility string, noOfPosts int) ([]*map[string]interface{}, error)
	ListFriendSuggestions(ctx context.Context, userId, startedThreshold string, noOfSuggestions int) ([]*map[string]string, error)
	UpdateMediaReaction(ctx context.Context, fromUserKey, toMediaKey, key string, newDoc map[string]interface{}) (string, error)
	ListOwnersPosts(ctx context.Context, userKey, lastPostTimestamp string, noOfPosts int) ([]*map[string]interface{}, error)
	CreateMediaReaction(ctx context.Context, fromUserKey, toMediaKey string, newDoc map[string]interface{}) (string, error)
	GetRole(authUserKey, userKey string) urepo.UserRole
	ListAllMedia(ctx context.Context, userKey string, offset, count int) ([]*map[string]string, error)
	ListPublicMedia(ctx context.Context, userKey string, offset, count int) ([]*map[string]string, error)
	GetUserKeyByIndexNo(ctx context.Context, indexNo string) (string, error)
}
