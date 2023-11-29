/**
 * This is a special type of repo which is not specific to a collection. This type of repo will use to execute more complex query
 *  on whole database, aggreate functionality of different collection (models) etc. Therefore, this is not bound to any collection.
 */

package repository

import (
	"context"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

type repo struct {
	db argdb.Interface
}

var _ Interface = (*repo)(nil)

func NewRepo(argIntfce argdb.Interface) *repo {
	return &repo{db: argIntfce}
}

func (r *repo) ListAllUsers(ctx context.Context) ([]*map[string]string, error) {
	return r.db.ListJsonStringValue(ctx, listAllUsersQry, map[string]interface{}{})
}
