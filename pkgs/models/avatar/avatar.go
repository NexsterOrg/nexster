package avatar

import (
	"context"
	"fmt"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

type avatarCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*avatarCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *avatarCtrler {
	return &avatarCtrler{argClient: argClient}
}

func (uc *avatarCtrler) MkAvatarDocId(key string) string {
	return fmt.Sprintf("%s/%s", AvatarColl, key)
}

func (ac *avatarCtrler) Get(ctx context.Context, key string) (*Avatar, error) {
	avatar := &Avatar{}
	_, err := ac.argClient.Coll.ReadDocument(ctx, key, avatar)
	return avatar, err
}
