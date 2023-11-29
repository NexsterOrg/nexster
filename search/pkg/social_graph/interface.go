package socialgraph

import "context"

type Interface interface {
	SearchAmongUsers(ctx context.Context, keyword string) ([]*map[string]string, error)
}
