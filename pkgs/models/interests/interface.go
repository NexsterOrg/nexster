package interests

import (
	"context"
)

const InterestsColl string = "interests"

// fields names
const expireAtField string = "expireAt"
const ytVideosField string = "ytVideos"

type Interface interface {
	ListExpiredInterests(ctx context.Context, limit int) ([]*InterestForExpiredList, error)
	RenewExpire(ctx context.Context, key string, newDate string) error
	StoreVidoes(ctx context.Context, key string, videos []*YoutubeVideo) error
	ListVideosForInterest(ctx context.Context, userKey string) ([]*YoutubeVideo, error)
}

type Interest struct {
	Key      string          `json:"_key"`
	Name     string          `json:"name"`
	ExpireAt string          `json:"expireAt"`
	YtVideos []*YoutubeVideo `json:"ytVideos"`
}

type YoutubeVideo struct {
	VId   string `json:"vId"`
	Title string `json:"title"`
	PubAt string `json:"pubAt"`
}

type InterestForExpiredList struct {
	Key  string `json:"_key"`
	Name string `json:"name"`
}
