package user

import (
	"context"
)

const UsersColl string = "users"

type Interface interface {
	MkUserDocId(key string) string
	ListUsers(ctx context.Context, query string, bindVars map[string]interface{}) ([]*User, error)
}

type User struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	ImageUrl string `json:"image_url"`
}
