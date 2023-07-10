package social_graph

import (
	"context"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

type Posts []*argdb.Media

type Users []*argdb.User

type Interface interface {
	ListRecentPosts(ctx context.Context, userNode, lastPostTimestamp, visibility string, noOfPosts int) (Posts, error)
	ListFriendSuggestions(ctx context.Context, userNode, startedThreshold string, noOfSuggestions int) (Users, error)
}
