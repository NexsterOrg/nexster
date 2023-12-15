package hasgender

import (
	"context"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	utud "github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
)

type hasGenderCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*hasGenderCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *hasGenderCtrler {
	return &hasGenderCtrler{argClient: argClient}
}

func (hgc *hasGenderCtrler) Create(ctx context.Context, data *HasGender) (string, error) {
	data.Key = utud.GenUUID4()
	data.Kind = hasGenderKind
	_, err := hgc.argClient.Coll.CreateDocument(ctx, data)
	if err != nil {
		return "", err
	}
	return data.Key, nil
}
