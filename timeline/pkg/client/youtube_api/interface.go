package youtubeapi

import (
	"context"

	intrs "github.com/NamalSanjaya/nexster/pkgs/models/interests"
)

type Interface interface {
	SearchYoutubeVideo(ctx context.Context, query, nextPgToken string, maxResults int) ([]*intrs.YoutubeVideo, error)
}

type SearchVideoResult struct {
	Id    string
	Title string
	PubAt string
}
