package socialgraph

import (
	"context"

	tp "github.com/NamalSanjaya/nexster/space/pkg/types"
)

type Interface interface {
	CreateEvent(ctx context.Context, userKey string, data *tp.Event) (string, string, error)
}
