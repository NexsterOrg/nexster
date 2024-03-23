package stemvideo

import (
	"context"
	"fmt"

	"github.com/NamalSanjaya/nexster/pkgs/cache/redis"
	errs "github.com/NamalSanjaya/nexster/pkgs/errors"
)

type StemVideoConfig struct {
	UserVideoFeedExpireTime int `yaml:"expireTime"` // user video expiration duration in minutes
}

type stemVideoCmd struct {
	redisClient redis.Interface
	config      *StemVideoConfig
}

var _ Interface = (*stemVideoCmd)(nil)

func New(cfg *StemVideoConfig, redisIntfce redis.Interface) *stemVideoCmd {
	return &stemVideoCmd{
		redisClient: redisIntfce,
		config:      cfg,
	}
}

// To create video cache key
func mkVideoKey(videoId string) string {
	return fmt.Sprintf("%s#%s", stemVideoBaseKey, videoId)
}

// To create video cache key
func mkVideoFeedKey(userKey string) string {
	return fmt.Sprintf("%s#%s", stemVideoFeedBaseKey, userKey)
}

func mkVideoFeedPropKey(userKey string) string {
	return fmt.Sprintf("%s#%s#status", stemVideoFeedPropBaseKey, userKey)
}

func (sv *stemVideoCmd) GetContent(ctx context.Context, videoId string) (*StemVideo, error) {
	content, err := sv.redisClient.GetHashFields(ctx, mkVideoKey(videoId), vId, title, publishedAt)
	if err != nil {
		return &StemVideo{}, err
	}
	if content[vId] == "" {
		return &StemVideo{}, errs.NewNotFoundError(fmt.Sprintf("hash not found for video id %s", videoId))
	}
	return &StemVideo{
		Id:          content[vId],
		Title:       content[title],
		PublishedAt: content[publishedAt],
	}, nil
}

func (sv *stemVideoCmd) ListAllVideoIdsForUser(ctx context.Context, userKey string) ([]string, error) {
	return sv.redisClient.LRangeAll(ctx, mkVideoFeedKey(userKey))
}

func (sv *stemVideoCmd) ListVideoIdsForUser(ctx context.Context, userKey string, startIndex, endIndex int) ([]string, error) {
	return sv.redisClient.LRange(ctx, mkVideoFeedKey(userKey), startIndex, endIndex)
}

func (sv *stemVideoCmd) IsFeedCreating(ctx context.Context, userKey string) (bool, error) {
	status, err := sv.redisClient.Get(ctx, mkVideoFeedPropKey(userKey))
	if errs.IsNotFoundError(err) {
		return false, nil
	}
	return status == statusCreating, err
}

func (sv *stemVideoCmd) StoreVideoIdsForUserFeed(ctx context.Context, userKey string, videoIds []string) error {
	key := mkVideoFeedKey(userKey)
	err := sv.redisClient.RPush(ctx, key, videoIds...)
	if err != nil {
		return err
	}
	return sv.redisClient.Expire(ctx, key, sv.config.UserVideoFeedExpireTime)
}

func (sv *stemVideoCmd) IsUserVideoFeedExist(ctx context.Context, userKey string) (bool, error) {
	return sv.redisClient.Exists(ctx, mkVideoFeedKey(userKey))
}
