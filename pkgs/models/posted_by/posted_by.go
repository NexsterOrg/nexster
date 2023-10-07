package postedby

import (
	"context"
	"fmt"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

type postedByCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*postedByCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *postedByCtrler {
	return &postedByCtrler{argClient: argClient}
}

func (pb *postedByCtrler) MkDocumentId(key string) string {
	return fmt.Sprintf("%s/%s", PostedByColl, key)
}

func (pb *postedByCtrler) CreateDocument(ctx context.Context, from, to, type1 string) (string, error) {
	meta, err := pb.argClient.Coll.CreateDocument(ctx, &PostedBy{
		From: from,
		To:   to,
		Type: type1,
		Kind: kind,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create postedBy edge document: %v", err)
	}
	return meta.Key, nil
}
