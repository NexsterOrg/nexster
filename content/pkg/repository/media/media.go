package media

import (
	"context"

	md "github.com/NamalSanjaya/nexster/pkgs/models/media"
)

type mediaRepository struct {
	ctrler md.Interface
}

var _ Interface = (*mediaRepository)(nil)

func NewRepository(mediaIntfce md.Interface) *mediaRepository {
	return &mediaRepository{
		ctrler: mediaIntfce,
	}
}

func (mr *mediaRepository) GetView(ctx context.Context, key string) (string, error) {
	data, err := mr.ctrler.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return data.Visibility, nil
}
