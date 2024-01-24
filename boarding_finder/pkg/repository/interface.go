package repository

import "context"

type Interface interface {
	ExistAndUniqueForMainContact(ctx context.Context, mainContact string) (bool, error)
}
