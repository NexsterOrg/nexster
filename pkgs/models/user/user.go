package user

import (
	"context"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

type userCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*userCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *userCtrler {
	return &userCtrler{argClient: argClient}
}

func (uc *userCtrler) ListUsers(ctx context.Context, query string, bindVars map[string]interface{}) ([]*User, error) {
	var users []*User
	cursor, err := uc.argClient.Db.Query(ctx, query, bindVars)
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

func (uc *userCtrler) MkUserDocId(key string) string {
	return fmt.Sprintf("%s/%s", UsersColl, key)
}

// To use outside places without User instance
func MkUserDocId(key string) string {
	return fmt.Sprintf("%s/%s", UsersColl, key)
}

// Return [{}, {}, {}]. json objects can have string type of values for fields.
func (uc *userCtrler) ListUsersV2(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]string, error) {
	results := []*map[string]string{}
	cursor, err := uc.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result map[string]string
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

// Return [{}, {}, {}]. json objects can have any type of values for fields.
func (uc *userCtrler) ListUsersAnyJsonValue(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]interface{}, error) {
	results := []*map[string]interface{}{}
	cursor, err := uc.argClient.Db.Query(ctx, query, bindVars)
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

// Return list of strings (eg: ["element1", "anotherElem2", "thirdElem3"] ). This will use when Query return value is list of strings.
func (uc *userCtrler) ListStrings(ctx context.Context, query string, bindVars map[string]interface{}) ([]*string, error) {
	results := []*string{}
	cursor, err := uc.argClient.Db.Query(ctx, query, bindVars)
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
		results = append(results, &result)
	}
}

func (uc *userCtrler) CountUsers(ctx context.Context, query string, bindVars map[string]interface{}) (int, error) {
	cursor, err := uc.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return 0, err
	}
	defer cursor.Close()

	for {
		var count int
		_, err := cursor.ReadDocument(ctx, &count)
		if driver.IsNoMoreDocuments(err) {
			return 0, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		return count, nil
	}
}

func (uc *userCtrler) GetUser(ctx context.Context, key string) (*User, error) {
	user := &User{}
	_, err := uc.argClient.Coll.ReadDocument(ctx, key, user)
	return user, err
}
