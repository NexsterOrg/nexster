package social_graph

import (
	"context"

	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
)

type Interface interface {
	ListRecentPosts(ctx context.Context, userId, lastPostTimestamp, visibility string, noOfPosts int) ([]*map[string]interface{}, error)
	ListFriendSuggestions(ctx context.Context, userId, startedThreshold string, noOfSuggestions int) ([]*urepo.User, error)
	UpdateMediaReaction(ctx context.Context, fromUserKey, toMediaKey, key string, newDoc map[string]interface{}) (string, error)
	ListOwnersPosts(ctx context.Context, userKey, lastPostTimestamp string, noOfPosts int) ([]*map[string]interface{}, error)
	CreateMediaReaction(ctx context.Context, fromUserKey, toMediaKey string, newDoc map[string]interface{}) (string, error)
}
