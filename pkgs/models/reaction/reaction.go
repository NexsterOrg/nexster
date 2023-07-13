package reaction

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

const ReactionColl string = "reactions"

const (
	like       string = "like"
	love       string = "love"
	laugh      string = "laugh"
	sad        string = "sad"
	insightful string = "insightful"
)

type reactionRepo struct {
	argClient *argdb.Client
}

var _ Interface = (*reactionRepo)(nil)

func NewRepo(argClient *argdb.Client) *reactionRepo {
	return &reactionRepo{argClient: argClient}
}

// This will update the document in Reaction collection for the given key.
// If the key is not existing, it will create a new document.
// fromUserId and toMediaId format should be "collection/key"
func (rerp *reactionRepo) UpdateReactions(ctx context.Context, fromUserId, toMediaId, key string, updateDoc map[string]interface{}) error {
	newDoc, err := convertBody(updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update reaction for id %s due to %v", key, err)
	}
	_, err = rerp.argClient.Coll.UpdateDocument(ctx, key, newDoc)
	if driver.IsArangoError(err) {
		// TODO
		// Key generation method should be placed here. do we need to generator a new key or go with given one??
		_, err = rerp.argClient.Coll.CreateDocument(ctx, createDocTemplate(fromUserId, toMediaId, key, newDoc))
		// Issue:
		// Edge is created even if the User or Media node is non-existing one.
		if err != nil {
			return fmt.Errorf("failed to create new doc %v", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to update doc %v", err)
	}
	return nil
}

func convertBody(doc map[string]interface{}) (map[string]bool, error) {
	newDoc := map[string]bool{}
	for key, val := range doc {
		if key == like || key == love || key == laugh || key == sad || key == insightful {
			temp, ok := val.(bool)
			if !ok {
				return newDoc, fmt.Errorf("invalid value name for key %s", key)
			}
			newDoc[key] = temp
			continue
		}
		return newDoc, fmt.Errorf("invalid key field %s", key)
	}
	return newDoc, nil
}

func createDocTemplate(from, to, key string, doc map[string]bool) *Reaction {
	newDoc := Reaction{
		Key:  key,
		From: from, // userId
		To:   to,   // mediaId
	}
	for key, value := range doc {
		switch key {
		case like:
			newDoc.Like = value
		case love:
			newDoc.Love = value
		case laugh:
			newDoc.Laugh = value
		case sad:
			newDoc.Sad = value
		case insightful:
			newDoc.Insightful = value
		}
	}
	return &newDoc
}
