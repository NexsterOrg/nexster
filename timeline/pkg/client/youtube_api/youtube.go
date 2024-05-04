package youtubeapi

import (
	"context"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	intrs "github.com/NamalSanjaya/nexster/pkgs/models/interests"
)

type YoutubeApi struct {
	client *youtube.Service
}

func NewClient(ctx context.Context, apiKey string) *YoutubeApi {
	ytClient, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		panic(err)
	}
	return &YoutubeApi{
		client: ytClient,
	}
}

func (yt *YoutubeApi) SearchYoutubeVideo(ctx context.Context, query, nextPgToken string, maxResults int) ([]*intrs.YoutubeVideo, error) {
	videos := []*intrs.YoutubeVideo{}
	call := yt.client.Search.List([]string{"snippet"}).PageToken(nextPgToken).VideoEmbeddable("true").
		Q(query).MaxResults(int64(maxResults)).SafeSearch("strict").Type("video").VideoDuration("short").VideoDuration("medium")

	// Execute the search request
	resp, err := call.Do()
	if err != nil {
		return videos, err
	}

	for _, item := range resp.Items {
		videos = append(videos, &intrs.YoutubeVideo{
			VId:   item.Id.VideoId,
			Title: item.Snippet.Title,
			PubAt: item.Snippet.PublishedAt,
		})
	}
	return videos, nil
}
