package user

import (
	"context"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	errs "github.com/NamalSanjaya/nexster/pkgs/errors"
	utuid "github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
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

// Create username for a user
func CreateUsername(firstName, secondName string) string {
	return fmt.Sprintf("%s %s", firstName, secondName)
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
	if driver.IsNotFoundGeneral(err) {
		return nil, errs.NewNotFoundError(fmt.Sprintf("document with key=%s is not found", key))
	}
	return user, err
}

func (uc *userCtrler) UpdateUser(ctx context.Context, key string, updateFields map[string]interface{}) error {
	_, err := uc.argClient.Coll.UpdateDocument(ctx, key, updateFields)
	if driver.IsNotFoundGeneral(err) {
		return errs.NewNotFoundError(fmt.Sprintf("document with key=%s is not found", key))
	}
	return err
}

func (uc *userCtrler) DeleteUser(ctx context.Context, key string) error {
	_, err := uc.argClient.Coll.RemoveDocument(ctx, key)
	if driver.IsNotFoundGeneral(err) {
		return errs.NewNotFoundError(fmt.Sprintf("document with key=%s is not found", key))
	}
	return err
}

func (uc *userCtrler) CreateDocument(ctx context.Context, doc *UserCreateInfo) (string, error) {
	doc.Key = utuid.GenUUID4()
	doc.Username = CreateUsername(doc.FirstName, doc.SecondName)

	_, err := uc.argClient.Coll.CreateDocument(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to create user node: %v", err)
	}
	return doc.Key, nil
}

func (uc *userCtrler) ListFacDepOfAllUsers(ctx context.Context) ([]*map[string]string, error) {
	return uc.ListUsersV2(ctx, listFacDepOfAllUsersQry, map[string]interface{}{})
}
