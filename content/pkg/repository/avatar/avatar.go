package avatar

import (
	"context"

	avtr "github.com/NamalSanjaya/nexster/pkgs/models/avatar"
)

type avatarRepo struct {
	ctrler avtr.Interface
}

var _ Interface = (*avatarRepo)(nil)

func NewRepo(intfce avtr.Interface) *avatarRepo {
	return &avatarRepo{
		ctrler: intfce,
	}
}

func (ar *avatarRepo) GetView(ctx context.Context, key string) (string, error) {
	data, err := ar.ctrler.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return data.View, nil
}
