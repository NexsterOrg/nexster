package repository

import (
	"context"
)

type Interface interface {
	ListUpcomingEvents(ctx context.Context, offset, count int) ([]*map[string]interface{}, error)
	GetEventReactionKey(ctx context.Context, userKey, eventKey string) (string, error)
	GetEvent(ctx context.Context, eventKey string) (map[string]interface{}, error)
	ListEventLovers(ctx context.Context, eventKey string, offset, count int) ([]*map[string]interface{}, error)
	GetEventOwnerKey(ctx context.Context, eventKey string) (string, error)
}
