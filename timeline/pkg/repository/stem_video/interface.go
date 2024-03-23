package stemvideo

import "context"

// base keys
const (
	stemVideoBaseKey         string = "stemVideo"
	stemVideoFeedBaseKey     string = "stemVideoFeed"
	stemVideoFeedPropBaseKey string = "stemVideoFeed#prop" // {stemVideoFeedBaseKey}#prop
)

// cache fields
const (
	vId         string = "vId" // video ID
	title       string = "title"
	publishedAt string = "pubAt"
)

// properties
const (
	statusCreating string = "creating"
)

type Interface interface {
	GetContent(ctx context.Context, videoId string) (*StemVideo, error)
	ListAllVideoIdsForUser(ctx context.Context, userKey string) ([]string, error)
	IsFeedCreating(ctx context.Context, userKey string) (bool, error)
	StoreVideoIdsForUserFeed(ctx context.Context, userKey string, videoIds []string) error
	ListVideoIdsForUser(ctx context.Context, userKey string, startIndex, endIndex int) ([]string, error)
	IsUserVideoFeedExist(ctx context.Context, userKey string) (bool, error)
}

type StemVideo struct {
	Id          string
	Title       string
	PublishedAt string
}
