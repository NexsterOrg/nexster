package repository

import (
	"context"
)

type Interface interface {
	ListUpcomingEvents(ctx context.Context, offset, count int) ([]*map[string]interface{}, error)
	GetEventReactionKey(ctx context.Context, userKey, eventKey string) (string, error)
}
