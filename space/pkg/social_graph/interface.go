package socialgraph

import (
	"context"

	"github.com/NamalSanjaya/nexster/pkgs/models/user"
	tp "github.com/NamalSanjaya/nexster/space/pkg/types"
)

type Interface interface {
	// return eventNodeKey, postedByEdgeKey, err
	CreateEvent(ctx context.Context, userKey string, data *tp.Event) (string, string, error)
	ListUpcomingEvents(ctx context.Context, userKey string, offset, count int) ([]*map[string]interface{}, error)
	GetEvent(ctx context.Context, userKey, eventKey string) (map[string]interface{}, error)
	ListEventReactUsersForType(ctx context.Context, eventKey, typ string, offset, count int) ([]*map[string]interface{}, error)
	GetEventOwnerKey(ctx context.Context, eventKey string) (string, error)
	GetRole(authUserKey, userKey string) user.UserRole
	CreateEventReactionEdge(ctx context.Context, reactorKey, eventKey string, data *tp.EventReaction) (string, error)
	SetEventReactionState(ctx context.Context, reactorKey, reactionEdgeKey string, data map[string]bool) error
}
