package redis

import (
	"context"
	"fmt"
	"time"

	rds "github.com/go-redis/redis/v8"

	errs "github.com/NamalSanjaya/nexster/pkgs/errors"
)

type Config struct {
	Host     string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	PassWord string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type client struct {
	client *rds.Client
}

var _ Interface = (*client)(nil)

func NewClient(ctx context.Context, config *Config) *client {
	rdsClient := rds.NewClient(&rds.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.PassWord,
		DB:       config.DB,
	})
	_, err := rdsClient.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Errorf("failed to connect redis server: %v", err))
	}
	return &client{
		client: rdsClient,
	}
}

func isRedisNil(err error) bool {
	return err == rds.Nil
}

func (rc *client) Exists(ctx context.Context, key string) (bool, error) {
	val, err := rc.client.Exists(ctx, key).Result()
	return val == 1, err
}

func (rc *client) Expire(ctx context.Context, key string, durationInMin int) error {
	return rc.client.Expire(ctx, key, time.Duration(durationInMin)*time.Minute).Err()
}

func (rc *client) Set(ctx context.Context, key, value string, expireTime int) error {
	return rc.client.Set(ctx, key, value, time.Duration(expireTime)).Err()
}

func (rc *client) SetIfNotExist(ctx context.Context, key, value string, expireTime int) error {
	return rc.client.SetNX(ctx, key, value, time.Duration(expireTime)).Err()
}

func (rc *client) Get(ctx context.Context, key string) (string, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if isRedisNil(err) {
		return "", errs.NewNotFoundError(fmt.Sprintf("key %s not found", key))
	}
	return val, nil
}

// append elements to a list
func (rc *client) RPush(ctx context.Context, key string, values ...string) error {
	return rc.client.RPush(ctx, key, values).Err()
}

// get all elements from a list
func (rc *client) LRangeAll(ctx context.Context, key string) ([]string, error) {
	return rc.client.LRange(ctx, key, 0, -1).Result()
}

func (rc *client) LRange(ctx context.Context, key string, startIndex, endIndex int) ([]string, error) {
	return rc.client.LRange(ctx, key, int64(startIndex), int64(endIndex)).Result()
}

func (rc *client) LIndex(ctx context.Context, key string, indx int) (string, error) {
	return rc.client.LIndex(ctx, key, int64(indx)).Result()
}

func (rc *client) SetHashFields(ctx context.Context, key string, values ...string) error {
	return rc.client.HSet(ctx, key, values).Err()
}

func (rc *client) SetHashFieldIfNotExist(ctx context.Context, key string, values ...string) error {
	isExist, err := rc.Exists(ctx, key)
	if err != nil {
		return err
	}
	if isExist {
		return nil
	}
	return rc.client.HSet(ctx, key, values).Err()
}

func (rc *client) GetHashField(ctx context.Context, key, field string) (string, error) {
	val, err := rc.client.HGet(ctx, key, field).Result()
	if isRedisNil(err) {
		return "", errs.NewNotFoundError(fmt.Sprintf("key %s or field %s is not found", key, field))
	}
	return val, nil
}

func (rc *client) DelHashFields(ctx context.Context, key string, fields ...string) error {
	return rc.client.HDel(ctx, key, fields...).Err()
}

func (rc *client) GetHashFields(ctx context.Context, key string, fields ...string) (map[string]string, error) {
	result := map[string]string{}
	values, err := rc.client.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return result, err
	}
	for i, val := range values {
		if val == nil {
			result[fields[i]] = ""
		} else {
			result[fields[i]] = val.(string)
		}
	}
	return result, nil
}

func (rc *client) HVals(ctx context.Context, key string) ([]string, error) {
	return rc.client.HVals(ctx, key).Result()
}

func (rc *client) ZRemRangeByScore(ctx context.Context, key, min, max string) error {
	return rc.client.ZRemRangeByScore(ctx, key, min, max).Err()
}

func (rc *client) ZRangeByScore(ctx context.Context, key string, minScore, maxScore string) ([]string, error) {
	arg := rds.ZRangeArgs{Key: key, Start: minScore, Stop: maxScore, ByScore: true}
	return rc.client.ZRangeArgs(ctx, arg).Result()
}

func (rc *client) ZRangeWithScore(ctx context.Context, key, min, max string,
	rev bool, offset, count int) ([]string, error) {
	return rc.client.ZRangeArgs(ctx, rds.ZRangeArgs{
		Key: key, Start: min, Stop: max, ByScore: true, ByLex: false,
		Rev: rev, Offset: int64(offset), Count: int64(count),
	}).Result()
}

func (rc *client) SMembers(ctx context.Context, key string) ([]string, error) {
	return rc.client.SMembers(ctx, key).Result()
}

func (rc *client) SRem(ctx context.Context, key string, values ...string) error {
	return rc.client.SRem(ctx, key, values).Err()
}

func (rc *client) SAdd(ctx context.Context, key string, value ...string) error {
	return rc.client.SAdd(ctx, key, value).Err()
}

// addding set of blockusers with uniqueness
func (rc *client) SSet(ctx context.Context, key string, value ...string) error {
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	if len(value) == 0 {
		return nil
	}
	return rc.client.SAdd(ctx, key, value).Err()
}
