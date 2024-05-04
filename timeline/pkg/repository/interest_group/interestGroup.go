package interestgroup

import (
	"context"
	"fmt"

	"github.com/NamalSanjaya/nexster/pkgs/cache/redis"
)

type interestGroupCmd struct {
	redisClient redis.Interface
}

var _ Interface = (*interestGroupCmd)(nil)

func New(redisIntfce redis.Interface) *interestGroupCmd {
	return &interestGroupCmd{
		redisClient: redisIntfce,
	}
}

func mkKey(groupId string) string {
	return fmt.Sprintf("%s#%s", interestGroupBaseKey, groupId)
}

func mkPropsStatusKey() string {
	return fmt.Sprintf("%s#status", interestGroupPropBaseKey)
}

func (igc *interestGroupCmd) ListVideoIdsForGroup(ctx context.Context, groupId string) ([]string, error) {
	return igc.redisClient.LRangeAll(ctx, mkKey(groupId))
}

func (igc *interestGroupCmd) IsCreating(ctx context.Context, groupId string) (bool, error) {
	status, err := igc.redisClient.Get(ctx, mkPropsStatusKey())
	return status == statusCreating, err
}
