package repository

import "context"

type Interface interface {
	ListAllUsers(ctx context.Context) ([]*map[string]string, error)
}
