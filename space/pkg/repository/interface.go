package repository

import (
	"context"
)

type Interface interface {
	ListUpcomingEvents(ctx context.Context, offset, count int) ([]*map[string]interface{}, error)
	GetEventReaction(ctx context.Context, userKey, eventKey string) (map[string]interface{}, error)
	GetEvent(ctx context.Context, eventKey string) (map[string]interface{}, error)
	ListEventLovers(ctx context.Context, eventKey string, offset, count int) ([]*map[string]interface{}, error)
	GetEventOwnerKey(ctx context.Context, eventKey string) (string, error)
	ListEventAttendees(ctx context.Context, eventKey string, offset, count int) ([]*map[string]interface{}, error)
	GetKeyOfUserReaction(ctx context.Context, eventKey, userKey string) (string, error)
	ListEventsForUser(ctx context.Context, userKey string, offset, count int) ([]*map[string]interface{}, error)
	DelPostedByGivenFromAndTo(ctx context.Context, fromId, toId string) error
}
