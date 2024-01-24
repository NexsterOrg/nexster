/**
 * This is a special type of repo which is not specific to a collection. This type of repo will use to execute more complex query
 * on whole database, aggreate functionality of different collection (models) etc. Therefore, this is not bound to any collection.
 */

package repository

import (
	"context"
	"fmt"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	"github.com/NamalSanjaya/nexster/pkgs/errors"
)

type repo struct {
	db argdb.Interface
}

var _ Interface = (*repo)(nil)

func NewRepo(argIntfce argdb.Interface) *repo {
	return &repo{db: argIntfce}
}

func (r *repo) ExistAndUniqueForMainContact(ctx context.Context, mainContact string) (bool, error) {
	bdOwnerKeys, err := r.db.ListStrings(ctx, occurrenceCountQry, map[string]interface{}{
		"mainContact": mainContact,
	})
	if err != nil {
		return false, err
	}
	ln := len(bdOwnerKeys)
	if ln == 0 {
		return false, nil
	}
	if ln == 1 {
		return true, nil
	}
	return false, errors.NewConflictError(fmt.Sprintf("more than one boarding owner keys exist: mainContact=%s", mainContact))
}
