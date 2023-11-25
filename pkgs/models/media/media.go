package media

import (
	"context"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	tm "github.com/NamalSanjaya/nexster/pkgs/utill/time"
	ud "github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
)

type mediaRepo struct {
	argClient *argdb.Client
}

var _ Interface = (*mediaRepo)(nil)

func NewRepo(argClient *argdb.Client) *mediaRepo {
	return &mediaRepo{argClient: argClient}
}

func (mr *mediaRepo) ListMedia(ctx context.Context, query string, bindVars map[string]interface{}) ([]*Media, error) {
	var medias []*Media
	cursor, err := mr.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return medias, err
	}
	defer cursor.Close()

	for {
		var media Media
		_, err := cursor.ReadDocument(ctx, &media)
		if driver.IsNoMoreDocuments(err) {
			return medias, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		medias = append(medias, &media)
	}
}

func (mr *mediaRepo) ListMediaWithOwner(ctx context.Context, query string, bindVars map[string]interface{}) ([]*MediaWithOwner, error) {
	var mediasWithOwners []*MediaWithOwner
	cursor, err := mr.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return mediasWithOwners, err
	}
	defer cursor.Close()

	for {
		var mediaWithOwner MediaWithOwner
		_, err := cursor.ReadDocument(ctx, &mediaWithOwner)
		if driver.IsNoMoreDocuments(err) {
			return mediasWithOwners, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		mediasWithOwners = append(mediasWithOwners, &mediaWithOwner)
	}
}

func (mr *mediaRepo) MkMediaDocId(key string) string {
	return fmt.Sprintf("%s/%s", MediaColl, key)
}

func (mr *mediaRepo) ListMediaWithCustomFields(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]string, error) {
	var medias []*map[string]string
	cursor, err := mr.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return medias, err
	}
	defer cursor.Close()

	for {
		var media map[string]string
		_, err := cursor.ReadDocument(ctx, &media)
		if driver.IsNoMoreDocuments(err) {
			return medias, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		medias = append(medias, &media)
	}
}

func (mr *mediaRepo) Get(ctx context.Context, key string) (*Media, error) {
	media := &Media{}
	_, err := mr.argClient.Coll.ReadDocument(ctx, key, media)
	return media, err
}

// Create media for given key. if key is empty string, this will create a new key
func (mr *mediaRepo) CreateForGivenKey(ctx context.Context, data *Media) (string, error) {
	data.Key = ud.GenUUID4()
	data.Size = 0 // NOTE: Setting size to zero, since we are not using it.
	data.CreateDate = tm.CurrentUTCTime()
	if _, err := mr.argClient.Coll.CreateDocument(ctx, data); err != nil {
		return "", err
	}
	return data.Key, nil
}

// Return list of strings (eg: ["element1", "anotherElem2", "thirdElem3"] ). This will use when Query return value is list of strings.
func (mr *mediaRepo) ListStrings(ctx context.Context, query string, bindVars map[string]interface{}) ([]string, error) {
	results := []string{}
	cursor, err := mr.argClient.Db.Query(ctx, query, bindVars)
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

func (mr *mediaRepo) DeleteDocument(ctx context.Context, mediaKey string) (map[string]interface{}, error) {
	result := &map[string]interface{}{}
	// to return deleted document
	ctx = driver.WithReturnOld(ctx, result)
	_, err := mr.argClient.Coll.RemoveDocument(ctx, mediaKey)
	return *result, err
}
