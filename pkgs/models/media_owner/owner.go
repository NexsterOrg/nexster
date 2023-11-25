package mediaowner

import (
	"context"
	"log"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	"github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
	driver "github.com/arangodb/go-driver"
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

func (moc *mediaOwnerCtrler) ListStringValueJson(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]string, error) {
	results := []*map[string]string{}
	cursor, err := moc.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		result := map[string]string{}
		_, err := cursor.ReadDocument(ctx, &result)
		if driver.IsNoMoreDocuments(err) {
			return results, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		results = append(results, &result)
	}
}
