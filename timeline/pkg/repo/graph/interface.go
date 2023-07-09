package graph

import (
	"context"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

type Interface interface {
	GetPostsForTimeline(ctx context.Context, userNode, lastPostTimestamp string, noOfPosts int) ([]*argdb.Media, error)
}
