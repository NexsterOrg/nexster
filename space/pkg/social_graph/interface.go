package socialgraph

import (
	"context"

	tp "github.com/NamalSanjaya/nexster/space/pkg/types"
)

type Interface interface {
	// return eventNodeKey, postedByEdgeKey, err
	CreateEvent(ctx context.Context, userKey string, data *tp.Event) (string, string, error)
	ListLatestEvents(ctx context.Context, offset, count int) ([]*map[string]string, error)
}
