package arangodb

import (
	"context"
)

// Interface related to Db client. Db client is able to execute query against whole arangodb database. Not specific to any collection.
type Interface interface {
	ListJsonAnyValue(ctx context.Context, query string, bindVar map[string]interface{}) ([]*map[string]interface{}, error)
	ListStrings(ctx context.Context, query string, bindVar map[string]interface{}) ([]string, error)
}
