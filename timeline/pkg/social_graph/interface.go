package social_graph

import (
	"context"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

const (
	UsersDoc string = "users"
)

type Posts []*argdb.Media

type Users []*argdb.User

type Interface interface {
	ListRecentPosts(ctx context.Context, userId, lastPostTimestamp, visibility string, noOfPosts int) (Posts, error)
	ListFriendSuggestions(ctx context.Context, userId, startedThreshold string, noOfSuggestions int) (Users, error)
}
