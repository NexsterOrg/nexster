package avatar

import "context"

type Interface interface {
	GetView(ctx context.Context, key string) (string, error)
}
