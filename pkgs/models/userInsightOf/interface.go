package userInsightOf

import "context"

const UserInsightOfColl string = "userInsightOf"

type Interface interface {
	MkDocumentId(key string) string
	Create(ctx context.Context, data *UserInsightOf) (string, error)
	Delete(ctx context.Context, key string) error
}

type UserInsightOf struct {
	Key  string `json:"_key"`
	From string `json:"_from"`
	To   string `json:"_to"`
}
