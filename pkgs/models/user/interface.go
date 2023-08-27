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
	ListStrings(ctx context.Context, query string, bindVars map[string]interface{}) ([]*string, error)
}

// TODO:
// User field has all user information. But some APIs don't need to
// fetch all user data. improve this
type DegreeInfo struct {
	Field string `json:"field"`
	Entry string `json:"entry"`
	End   string `json:"end"`
}

type User struct {
	UserId   string `json:"_key"`
	Username string `json:"username"`
	ImageUrl string `json:"image_url"`
	Headling string `json:"headling"`
	Faculty  string `json:"faculty"`
	Field    string `json:"field"`
	Batch    string `json:"batch"`
	About    string `json:"about"`
}

type UserRole int

const (
	Owner UserRole = iota
	Viewer
)
