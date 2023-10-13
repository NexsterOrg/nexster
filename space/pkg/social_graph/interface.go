package socialgraph

import (
	"context"

	tp "github.com/NamalSanjaya/nexster/space/pkg/types"
)

type Interface interface {
	// return eventNodeKey, postedByEdgeKey, err
	CreateEvent(ctx context.Context, userKey string, data *tp.Event) (string, string, error)
	ListUpcomingEvents(ctx context.Context, userKey string, offset, count int) ([]*map[string]interface{}, error)
}
