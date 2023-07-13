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
	var posts []*Media
	cursor, err := mr.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return posts, err
	}
	defer cursor.Close()

	for {
		var post Media
		_, err := cursor.ReadDocument(ctx, &post)
		if driver.IsNoMoreDocuments(err) {
			return posts, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		posts = append(posts, &post)
	}
}

func (mr *mediaRepo) MkMediaDocId(key string) string {
	return fmt.Sprintf("%s/%s", MediaColl, key)
}
