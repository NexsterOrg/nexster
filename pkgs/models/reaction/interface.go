package reaction

import "context"

type Interface interface {
	UpdateReactions(ctx context.Context, fromUserId, toMediaId, key string, updateDoc map[string]interface{}) (string, error)
	GetReactionsCount(ctx context.Context, query string, bindVars map[string]interface{}) (map[string]int, error)
	CreateReactionLink(ctx context.Context, fromUserId, toMediaId string, updateDoc map[string]interface{}) (string, error)
}

type Reaction struct {
	From       string `json:"_from,omitempty"`
	To         string `json:"_to,omitempty"`
	Key        string `json:"_key,omitempty"`
	Like       bool   `json:"like"`
	Love       bool   `json:"love"`
	Laugh      bool   `json:"laugh"`
	Sad        bool   `json:"sad"`
	Insightful bool   `json:"insightful"`
}
