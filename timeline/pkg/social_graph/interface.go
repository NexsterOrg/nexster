package social_graph

import (
	"context"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

type Posts []*argdb.Media

type Interface interface {
	ListRecentPosts(ctx context.Context, userNode, lastPostTimestamp, visibility string, noOfPosts int) (Posts, error)
}
