package student

import (
	"context"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	utud "github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
)

type studentCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*studentCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *studentCtrler {
	return &studentCtrler{argClient: argClient}
}

func (ac *studentCtrler) Create(ctx context.Context, doc *Student) (string, error) {
	doc.Key = utud.GenUUID4()
	doc.Kind = StudnetColl
	_, err := ac.argClient.Coll.CreateDocument(ctx, doc)
	if err != nil {
		return "", err
	}
	return doc.Key, nil
}
