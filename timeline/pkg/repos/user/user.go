package user

import (
	"context"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

type userRepo struct {
	argClient *argdb.Client
}

var _ Interface = (*userRepo)(nil)

func NewRepo(argClient *argdb.Client) *userRepo {
	return &userRepo{argClient: argClient}
}

func (ur *userRepo) ListUsers(ctx context.Context, query string, bindVars map[string]interface{}) ([]*User, error) {
	var users []*User
	cursor, err := ur.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return users, err
	}
	defer cursor.Close()

	for {
		var user User
		_, err := cursor.ReadDocument(ctx, &user)
		if driver.IsNoMoreDocuments(err) {
			return users, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		users = append(users, &user)
	}
}

func (ur *userRepo) MkUserDocId(key string) string {
	return fmt.Sprintf("%s/%s", UsersColl, key)
}
