package mediaowner

import (
	"context"
	"log"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	errs "github.com/NamalSanjaya/nexster/pkgs/errors"
	md "github.com/NamalSanjaya/nexster/pkgs/models/media"
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

// Return list of strings (eg: ["element1", "anotherElem2", "thirdElem3"] ). This will use when Query return value is list of strings.
func (moc *mediaOwnerCtrler) listStrings(ctx context.Context, query string, bindVars map[string]interface{}) ([]string, error) {
	results := []string{}
	cursor, err := moc.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result string
		_, err := cursor.ReadDocument(ctx, &result)
		if driver.IsNoMoreDocuments(err) {
			return results, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		results = append(results, result)
	}
}

func (moc *mediaOwnerCtrler) GetOwnerForMedia(ctx context.Context, mediaKey string) (string, error) {
	resList, err := moc.listStrings(ctx, getOwnerForMediaQry, map[string]interface{}{
		"mediaNode": md.MkMediaDocumentId(mediaKey),
	})
	if err != nil {
		return "", err
	}
	if len(resList) == 0 {
		return "", errs.NewNotFoundError("document not found")
	}
	return resList[0], nil
}
