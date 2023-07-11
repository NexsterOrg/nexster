package social_graph

import (
	"context"

	mrepo "github.com/NamalSanjaya/nexster/timeline/pkg/repos/media"
	urepo "github.com/NamalSanjaya/nexster/timeline/pkg/repos/user"
)

type Interface interface {
	ListRecentPosts(ctx context.Context, userId, lastPostTimestamp, visibility string, noOfPosts int) ([]*mrepo.Media, error)
	ListFriendSuggestions(ctx context.Context, userId, startedThreshold string, noOfSuggestions int) ([]*urepo.User, error)
	UpdateMediaReaction(ctx context.Context, fromUserKey, toMediaKey, key string, newDoc map[string]interface{}) error
}
