package interestsin

import (
	"context"
	"fmt"
	"log"

	"github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	errs "github.com/NamalSanjaya/nexster/pkgs/errors"
	ig "github.com/NamalSanjaya/nexster/pkgs/models/interestGroups"
	usr "github.com/NamalSanjaya/nexster/pkgs/models/user"
	utud "github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
)

type interestsInCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*interestsInCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *interestsInCtrler {
	return &interestsInCtrler{argClient: argClient}
}

func (iic *interestsInCtrler) Create(ctx context.Context, fromUserKey, toInterestGroupKey string) (string, error) {
	meta, err := iic.argClient.Coll.CreateDocument(ctx, &InterestsIn{
		Key:  utud.GenUUID4(),
		Kind: interestsInKind,
		From: usr.MkUserDocId(fromUserKey),
		To:   ig.MkInterestGroupDocId(toInterestGroupKey),
	})
	if err != nil {
		return "", err
	}
	return meta.Key, nil
}

func (iic *interestsInCtrler) Delete(ctx context.Context, key string) error {
	_, err := iic.argClient.Coll.RemoveDocument(ctx, key)
	if driver.IsNotFoundGeneral(err) {
		return errs.NewNotFoundError(fmt.Sprintf("document, key=%s is not found", key))
	}
	return err
}

// Return list of strings (eg: ["element1", "anotherElem2", "thirdElem3"] ). This will use when Query return value is list of strings.
func (iic *interestsInCtrler) listStrings(ctx context.Context, query string, bindVars map[string]interface{}) ([]string, error) {
	results := []string{}
	cursor, err := iic.argClient.Db.Query(ctx, query, bindVars)
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

func (iic *interestsInCtrler) DeleteByFromTo(ctx context.Context, fromUserKey, toInterestGroupKey string) error {
	_, err := iic.listStrings(ctx, deleteByFromToQry, map[string]interface{}{
		"from": usr.MkUserDocId(fromUserKey),
		"to":   ig.MkInterestGroupDocId(toInterestGroupKey),
	})
	return err
}

func (iic *interestsInCtrler) InsertByFacDepName(ctx context.Context, facDepName, fromUserKey string) error {
	_, err := iic.listStrings(ctx, insertDocByFacDepName, map[string]interface{}{
		"facDepName": facDepName,
		"userNode":   usr.MkUserDocId(fromUserKey),
		"kind":       interestsInKind,
	})
	return err
}
