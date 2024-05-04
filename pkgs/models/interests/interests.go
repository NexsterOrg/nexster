package interests

import (
	"context"
	"fmt"
	"log"

	"github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	errs "github.com/NamalSanjaya/nexster/pkgs/errors"
	usr "github.com/NamalSanjaya/nexster/pkgs/models/user"
)

type interestsCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*interestsCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *interestsCtrler {
	return &interestsCtrler{argClient: argClient}
}

func MkDocumentId(key string) string {
	return fmt.Sprintf("%s/%s", InterestsColl, key)
}

// List `limit` no.of interests which are needed to be renewed.
func (ic *interestsCtrler) ListExpiredInterests(ctx context.Context, limit int) ([]*InterestForExpiredList, error) {
	results := []*InterestForExpiredList{}
	cursor, err := ic.argClient.Db.Query(ctx, listBasicInterestInfoQry, map[string]interface{}{
		"limit": limit,
	})
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result InterestForExpiredList
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

func (ic *interestsCtrler) update(ctx context.Context, key string, updateFields map[string]interface{}) error {
	_, err := ic.argClient.Coll.UpdateDocument(ctx, key, updateFields)
	if driver.IsNotFoundGeneral(err) {
		return errs.NewNotFoundError(fmt.Sprintf("document with key=%s is not found", key))
	}
	return err
}

func (ic *interestsCtrler) RenewExpire(ctx context.Context, key string, newDate string) error {
	return ic.update(ctx, key, map[string]interface{}{
		expireAtField: newDate,
	})
}

func (ic *interestsCtrler) StoreVidoes(ctx context.Context, key string, videos []*YoutubeVideo) error {
	return ic.update(ctx, key, map[string]interface{}{
		ytVideosField: videos,
	})
}

func (ic *interestsCtrler) ListVideosForInterest(ctx context.Context, userKey string) ([]*YoutubeVideo, error) {
	results := new([]*YoutubeVideo)
	cursor, err := ic.argClient.Db.Query(ctx, listYtVideosForInterestQry, map[string]interface{}{
		"userNode": usr.MkUserDocId(userKey),
	})
	if err != nil {
		return *results, err
	}
	defer cursor.Close()

	for {
		result := []*YoutubeVideo{}
		_, err := cursor.ReadDocument(ctx, &result)
		if driver.IsNoMoreDocuments(err) {
			return *results, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		*results = append(*results, result...)
	}
}
