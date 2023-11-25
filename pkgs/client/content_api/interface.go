package contentapi

import (
	"context"
)

// permission types
const (
	Owner  string = "owner"
	Viewer string = "viewer"
)

type Interface interface {
	CreateImageUrl(imgIdWithNamespace, permission string) (string, error)
	GetPermission(ownerKey, viewerKey string) string
	DeleteImage(ctx context.Context, imgIdWithNamespace string) error
}
