package media

import (
	"context"
)

type Interface interface {
	GetView(ctx context.Context, key string) (string, error)
	GetViewForLink(ctx context.Context, link string) (string, error)
}
