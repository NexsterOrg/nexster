package avatar

import (
	"context"
)

const AvatarColl string = "avatars"

type Avatar struct {
	Key       string `json:"_key"`
	Namespace string `json:"namespace"`
	Format    string `json:"format"`
	View      string `json:"view"`
}

type Interface interface {
	MkAvatarDocId(key string) string
	Get(ctx context.Context, key string) (*Avatar, error)
}
