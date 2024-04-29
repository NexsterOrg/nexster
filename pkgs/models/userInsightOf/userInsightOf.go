package userInsightOf

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	er "github.com/NamalSanjaya/nexster/pkgs/errors"
	utud "github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
)

type userInsightOfCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*userInsightOfCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *userInsightOfCtrler {
	return &userInsightOfCtrler{argClient: argClient}
}

func (uioc *userInsightOfCtrler) MkDocumentId(key string) string {
	return fmt.Sprintf("%s/%s", UserInsightOfColl, key)
}

// To use outside places without User Insight Of instance
func MkDocumentId(key string) string {
	return fmt.Sprintf("%s/%s", UserInsightOfColl, key)
}

func (uio *userInsightOfCtrler) Create(ctx context.Context, data *UserInsightOf) (string, error) {
	data.Key = utud.GenUUID4()
	_, err := uio.argClient.Coll.CreateDocument(ctx, data)
	if err != nil {
		return "", err
	}
	return data.Key, nil
}

func (uio *userInsightOfCtrler) Delete(ctx context.Context, key string) error {
	_, err := uio.argClient.Coll.RemoveDocument(ctx, key)
	if driver.IsNotFoundGeneral(err) {
		return er.NewNotFoundError(fmt.Sprintf("document with key=%s is not found", key))
	}
	return err
}
