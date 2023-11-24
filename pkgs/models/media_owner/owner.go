package mediaowner

import (
	"context"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	"github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
)

type mediaOwnerCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*mediaOwnerCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *mediaOwnerCtrler {
	return &mediaOwnerCtrler{argClient: argClient}
}

func (moc *mediaOwnerCtrler) Create(ctx context.Context, fromId, toId string) (string, error) {
	key := uuid.GenUUID4()
	if _, err := moc.argClient.Coll.CreateDocument(ctx, &MediaOwner{
		Key: key, Kind: MediaOwnerKind, From: fromId, To: toId,
	}); err != nil {
		return "", err
	}
	return key, nil
}
