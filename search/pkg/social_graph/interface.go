package socialgraph

import "context"

type Interface interface {
	SearchAmongUsers(ctx context.Context, keyword string) ([]*map[string]string, error)
	AttachFriendState(ctx context.Context, reqstorKey, friendKey string) (state string, reqId string, err error)
}
