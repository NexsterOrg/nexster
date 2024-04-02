package hasgender

import (
	"context"
	"log"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	gnd "github.com/NamalSanjaya/nexster/pkgs/models/genders"
	usr "github.com/NamalSanjaya/nexster/pkgs/models/user"
	utud "github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
	"github.com/arangodb/go-driver"
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

// Return list of strings (eg: ["element1", "anotherElem2", "thirdElem3"] ). This will use when Query return value is list of strings.
func (hgc *hasGenderCtrler) listStrings(ctx context.Context, query string, bindVars map[string]interface{}) ([]string, error) {
	results := []string{}
	cursor, err := hgc.argClient.Db.Query(ctx, query, bindVars)
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

func (hgc *hasGenderCtrler) DeleteByFromTo(ctx context.Context, fromUserKey, toGenderKey string) error {
	_, err := hgc.listStrings(ctx, deleteByFromToQry, map[string]interface{}{
		"from": usr.MkUserDocId(fromUserKey),
		"to":   gnd.MkGenderId(toGenderKey),
	})
	return err
}
