/**
 * This is a special type of repo which is not specific to a collection. This type of repo will use to execute more complex query
 *  on whole database, aggreate functionality of different collection (models) etc. Therefore, this is not bound to any collection.
 */

package repository

import (
	"context"
	"fmt"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	"github.com/NamalSanjaya/nexster/pkgs/errors"
	"github.com/NamalSanjaya/nexster/pkgs/models/event"
	"github.com/NamalSanjaya/nexster/pkgs/models/user"
)

type repo struct {
	db argdb.Interface
}

var _ Interface = (*repo)(nil)

func NewRepo(argIntfce argdb.Interface) *repo {
	return &repo{db: argIntfce}
}

func (r *repo) ListUpcomingEvents(ctx context.Context, offset, count int) ([]*map[string]interface{}, error) {
	return r.db.ListJsonAnyValue(ctx, upcomingEventsQry, map[string]interface{}{
		"offset": offset,
		"count":  count,
	})
}

func (r *repo) GetEventReaction(ctx context.Context, userKey, eventKey string) (map[string]interface{}, error) {
	emptyResult := map[string]interface{}{
		"key":   "",
		"love":  false,
		"going": false,
	}
	keys, err := r.db.ListJsonAnyValue(ctx, getEventReactionKeyQry, map[string]interface{}{
		"userNode":  user.MkUserDocId(userKey),
		"eventNode": event.MkEventDocId(eventKey),
	})
	if err != nil {
		return emptyResult, err
	}
	ln := len(keys)
	if ln > 1 {
		return emptyResult, fmt.Errorf("more than one event reaction edges exist from=%s to=%s", userKey, eventKey)
	}
	if ln == 0 {
		return emptyResult, nil
	}
	return *(keys[0]), nil
}

func (r *repo) GetEvent(ctx context.Context, eventKey string) (map[string]interface{}, error) {
	emptyResult := map[string]interface{}{}
	event, err := r.db.ListJsonAnyValue(ctx, getEventQry, map[string]interface{}{
		"eventNode": event.MkEventDocId(eventKey),
	})
	if err != nil {
		return emptyResult, err
	}
	ln := len(event)
	if ln == 0 {
		return emptyResult, errors.NewNotFoundError("event is not found")
	}
	if ln > 1 {
		return emptyResult, fmt.Errorf("found more than one event nodes for eventKey=%s", eventKey)
	}
	return *event[0], nil
}

func (r *repo) ListEventLovers(ctx context.Context, eventKey string, offset, count int) ([]*map[string]interface{}, error) {
	return r.db.ListJsonAnyValue(ctx, getEventLoveUserQry, map[string]interface{}{
		"eventNode": event.MkEventDocId(eventKey),
		"offset":    offset,
		"count":     count,
	})
}

func (r *repo) ListEventAttendees(ctx context.Context, eventKey string, offset, count int) ([]*map[string]interface{}, error) {
	return r.db.ListJsonAnyValue(ctx, getEventGoingUserQry, map[string]interface{}{
		"eventNode": event.MkEventDocId(eventKey),
		"offset":    offset,
		"count":     count,
	})
}

func (r *repo) GetEventOwnerKey(ctx context.Context, eventKey string) (string, error) {
	ownerKeys, err := r.db.ListStrings(ctx, getOwnerUserKey, map[string]interface{}{
		"eventNode": event.MkEventDocId(eventKey),
	})
	if err != nil {
		return "", err
	}
	ln := len(ownerKeys)
	if ln > 1 {
		return "", fmt.Errorf("more than one owner keys exist: eventKey=%s", eventKey)
	}
	if ln == 0 {
		return "", nil
	}
	return ownerKeys[0], nil
}

func (r *repo) GetKeyOfUserReaction(ctx context.Context, eventKey, userKey string) (string, error) {
	eventReactEdgeKeys, err := r.db.ListStrings(ctx, getEventEdgeKeyOfUserReaction, map[string]interface{}{
		"eventNode": event.MkEventDocId(eventKey),
		"userNode":  user.MkUserDocId(userKey),
	})
	if err != nil {
		return "", err
	}
	ln := len(eventReactEdgeKeys)
	if ln > 1 {
		return "", fmt.Errorf("more than one event reaction edge keys exist: eventKey=%s, userKey=%s", eventKey, userKey)
	}
	if ln == 0 {
		return "", nil
	}
	return eventReactEdgeKeys[0], nil
}
