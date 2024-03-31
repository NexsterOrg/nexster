package interestsin

import "context"

const InterestsInColl string = "interestsIn"

const interestsInKind string = "interestsIn"

type Interface interface {
	Create(ctx context.Context, fromUserKey, toInterestGroupKey string) (string, error)
	Delete(ctx context.Context, key string) error
	DeleteByFromTo(ctx context.Context, fromUserKey, toInterestGroupKey string) error
	InsertByFacDepName(ctx context.Context, facDepName, fromUserKey string) error
}

type InterestsIn struct {
	Key  string `json:"_key"`
	From string `json:"_from"`
	To   string `json:"_to"`
	Kind string `json:"kind"`
}
