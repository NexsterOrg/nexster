package user

import (
	"context"
)

const UsersColl string = "users"

type Interface interface {
	MkUserDocId(key string) string
	ListUsers(ctx context.Context, query string, bindVars map[string]interface{}) ([]*User, error)
	ListUsersV2(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]string, error)
	CountUsers(ctx context.Context, query string, bindVars map[string]interface{}) (int, error)
	GetUser(ctx context.Context, key string) (*User, error)
}

type User struct {
	UserId   string `json:"_key"`
	Username string `json:"username"`
	ImageUrl string `json:"image_url"`
	Headling string `json:"headling"`
}
