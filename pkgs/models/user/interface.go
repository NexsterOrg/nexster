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
	ListUsersAnyJsonValue(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]interface{}, error)
	UpdateUser(ctx context.Context, key string, updateFields map[string]interface{}) error
	DeleteUser(ctx context.Context, key string) error
	CreateDocument(ctx context.Context, doc *UserCreateInfo) (string, error)
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
	UserId     string `json:"_key"`
	Username   string `json:"username"`
	ImageUrl   string `json:"image_url"`
	Faculty    string `json:"faculty"`
	Field      string `json:"field"`
	Batch      string `json:"batch"`
	About      string `json:"about"`
	FirstName  string `json:"firstName"`
	SecondName string `json:"secondName"`
	Gender     string `json:"gender"`
	Birthday   string `json:"birthday"`
	IndexNo    string `json:"index_no"`
	LastLogin  string `json:"lastLogin"`
}

// username, email

type UserCreateInfo struct {
	Key        string   `json:"_key"`
	FirstName  string   `json:"firstName"`
	SecondName string   `json:"secondName"`
	Username   string   `json:"username"`
	IndexNo    string   `json:"index_no"`
	Email      string   `json:"email"`
	ImageUrl   string   `json:"image_url"`
	Birthday   string   `json:"birthday"`
	Faculty    string   `json:"faculty"`
	Field      string   `json:"field"`
	Batch      string   `json:"batch"`
	About      string   `json:"about"`
	Gender     string   `json:"gender"`
	Password   string   `json:"password"`
	Roles      []string `json:"roles"`
}

type UserRole int

const (
	Owner UserRole = iota
	Viewer
)
