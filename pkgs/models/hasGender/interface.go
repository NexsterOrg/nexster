package hasgender

import "context"

const HasGenderColl string = "hasGender"
const hasGenderKind string = "has_gender"

type Interface interface {
	Create(ctx context.Context, data *HasGender) (string, error)
	DeleteByFromTo(ctx context.Context, fromUserKey, toGenderKey string) error
}

type HasGender struct {
	Key  string `json:"_key"`
	From string `json:"_from"`
	To   string `json:"_to"`
	Kind string `json:"kind"`
}
