package userInsight

import (
	"context"
)

const UserInsightsColl string = "userInsights"

type Interface interface {
	MkUserInsightDocId(key string) string
	CreateUserInsight(ctx context.Context, insightData *InsightData) (string, error)
	GetUserInsight(ctx context.Context, key string) (*InsightData, error)
	UpdateUserInsight(ctx context.Context, key string, updateFields map[string]interface{}) error
	DeleteUserInsight(ctx context.Context, key string) error
	CountUsers(ctx context.Context, query string, bindVars map[string]interface{}) (int, error)
	AppendLoginTimestamp(ctx context.Context, key, timestamp string) error
	GetActiveUserCountForGivenTimeRange(ctx context.Context, from, to string) (int, error)
}

type InsightData struct {
	Key             string   `json:"_key"`
	Type            string   `json:"type"`
	UserId          string   `json:"userId"`
	Year            string   `json:"year"`
	LoginTimestamps []string `json:"loginTimestamps"`
}
