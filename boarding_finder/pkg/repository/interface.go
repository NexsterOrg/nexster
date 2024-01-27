package repository

import "context"

type Interface interface {
	ExistAndUniqueForMainContact(ctx context.Context, mainContact string) (bool, error)
	IsUniqueEdgeExist(ctx context.Context, fromId, toId string) (bool, error)
	DelEdgeFromTo(ctx context.Context, fromId, toId string) error
}
