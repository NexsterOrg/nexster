package boardingowner

import (
	"context"
	"fmt"
	"log"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	tm "github.com/NamalSanjaya/nexster/pkgs/utill/time"
	ud "github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
	"github.com/arangodb/go-driver"
)

type boardingOwnerCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*boardingOwnerCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *boardingOwnerCtrler {
	return &boardingOwnerCtrler{argClient: argClient}
}

func MkDocId(key string) string {
	return fmt.Sprintf("%s/%s", BdOwnerColl, key)
}

func (bac *boardingOwnerCtrler) Create(ctx context.Context, doc *BoardingOwner) (key string, err error) {
	doc.Key = ud.GenUUID4()
	doc.CreatedAt = tm.CurrentUTCTime()
	_, err = bac.argClient.Coll.CreateDocument(ctx, doc)
	if err != nil {
		key = ""
		return
	}
	key = doc.Key
	return
}

func (bac *boardingOwnerCtrler) Exist(ctx context.Context, key string) (bool, error) {
	return bac.argClient.Coll.DocumentExists(ctx, key)
}

// Return [{}, {}, {}]. json objects can have any type of values for fields.
func (bac *boardingOwnerCtrler) ListAnyJsonValue(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]interface{}, error) {
	results := []*map[string]interface{}{}
	cursor, err := bac.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result map[string]interface{}
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
