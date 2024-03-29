package redis

import (
	"context"
)

type Interface interface {
	Expire(ctx context.Context, key string, durationInMin int) error
	Exists(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key, value string, expireTime int) error
	SetIfNotExist(ctx context.Context, key, value string, expireTime int) error
	Get(ctx context.Context, key string) (string, error)
	RPush(ctx context.Context, key string, values ...string) error
	LRangeAll(ctx context.Context, key string) ([]string, error)
	LRange(ctx context.Context, key string, startIndex, endIndex int) ([]string, error)
	LIndex(ctx context.Context, key string, indx int) (string, error)
	SetHashFields(ctx context.Context, key string, values ...string) error
	SetHashFieldIfNotExist(ctx context.Context, key string, values ...string) error
	GetHashField(ctx context.Context, key, field string) (string, error)
	DelHashFields(ctx context.Context, key string, fields ...string) error
	GetHashFields(ctx context.Context, key string, fields ...string) (map[string]string, error)
	HVals(ctx context.Context, key string) ([]string, error)
	ZRemRangeByScore(ctx context.Context, key, min, max string) error
	ZRangeByScore(ctx context.Context, key string, minScore, maxScore string) ([]string, error)
	ZRangeWithScore(ctx context.Context, key, min, max string, rev bool, offset, count int) ([]string, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	SRem(ctx context.Context, key string, values ...string) error
	SAdd(ctx context.Context, key string, value ...string) error
	SSet(ctx context.Context, key string, value ...string) error
}
