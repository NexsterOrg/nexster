package boardingadowned

import (
	"context"
	"fmt"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	"github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
)

type bdAdOwnedCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*bdAdOwnedCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *bdAdOwnedCtrler {
	return &bdAdOwnedCtrler{argClient: argClient}
}

func MkDocumentId(key string) string {
	return fmt.Sprintf("%s/%s", BdAdOwnedColl, key)
}

func (b *bdAdOwnedCtrler) CreateDocument(ctx context.Context, from, to string) (string, error) {
	meta, err := b.argClient.Coll.CreateDocument(ctx, &BdAdOwned{
		Key:  uuid.GenUUID4(),
		From: from,
		To:   to,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create boarding owned edge document: %v", err)
	}
	return meta.Key, nil
}
