package boardingads

import (
	"context"
	"fmt"
	"log"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	er "github.com/NamalSanjaya/nexster/pkgs/errors"
	tm "github.com/NamalSanjaya/nexster/pkgs/utill/time"
	ud "github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
	"github.com/arangodb/go-driver"
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

// results format [ {}, {}, {} ]
func (bac *bdAdsCtrler) getAdWithOwner(ctx context.Context, adId string) ([]*BdAdsWithOwner, error) {
	results := []*BdAdsWithOwner{}
	cursor, err := bac.argClient.Db.Query(ctx, getAdWithOwnerQry, map[string]interface{}{
		"adId": adId,
	})
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result BdAdsWithOwner
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

func (bac *bdAdsCtrler) GetAdWithOwner(ctx context.Context, adId string) (result *BdAdsWithOwner, err error) {
	result = &BdAdsWithOwner{}
	results, err := bac.getAdWithOwner(ctx, adId)
	if err != nil {
		return
	}
	ln := len(results)
	if ln == 0 {
		err = er.NewNotFoundError("ad not found")
		return
	}
	if ln > 1 {
		err = er.NewConflictError("more than one ads found")
		return
	}
	result = results[0]
	return
}

func (bac *bdAdsCtrler) Update(ctx context.Context, key string, updateFields map[string]interface{}) error {
	_, err := bac.argClient.Coll.UpdateDocument(ctx, key, updateFields)
	if driver.IsNotFoundGeneral(err) {
		return er.NewNotFoundError(fmt.Sprintf("document with key=%s is not found", key))
	}
	return err
}

func (bac *bdAdsCtrler) ListAdsWithFilters(ctx context.Context, minRent, maxRent, maxDist, minBeds, maxBeds, minBaths, maxBaths,
	offset, count int, sortBy string, genders, billTypes []string) ([]*AdInfoForList, error) {
	results := []*AdInfoForList{}
	var query string
	if sortBy == "rental" {
		query = listAdsSortByRental
	} else {
		query = listAdsSortByDate
	}
	cursor, err := bac.argClient.Db.Query(ctx, query, map[string]interface{}{
		"status":      Accepted,
		"minRent":     minRent,
		"maxRent":     maxRent,
		"maxDistance": maxDist,
		"minBeds":     minBeds,
		"maxBeds":     maxBeds,
		"minBaths":    minBaths,
		"maxBaths":    maxBaths,
		"genders":     genders,
		"billTypes":   billTypes,
		"offset":      offset,
		"count":       count,
	})
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result AdInfoForList
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
