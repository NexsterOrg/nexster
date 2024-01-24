package boardingads

import (
	"context"
	"fmt"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	tm "github.com/NamalSanjaya/nexster/pkgs/utill/time"
	ud "github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
)

type bdAdsCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*bdAdsCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *bdAdsCtrler {
	return &bdAdsCtrler{argClient: argClient}
}

func MkBdAdsDocId(key string) string {
	return fmt.Sprintf("%s/%s", BdAdsColl, key)
}

// return: createdDocKey, err
func (bac *bdAdsCtrler) Create(ctx context.Context, doc *BoardingAds) (string, error) {
	doc.Key = ud.GenUUID4()
	doc.CreatedAt = tm.CurrentUTCTime()
	_, err := bac.argClient.Coll.CreateDocument(ctx, doc)
	if err != nil {
		return "", err
	}
	return doc.Key, nil
}
