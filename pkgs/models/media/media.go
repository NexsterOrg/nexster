package media

import (
	"context"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
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
