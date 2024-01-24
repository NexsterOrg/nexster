package boardingadowned

import (
	"context"
)

const BdAdOwnedColl string = "boardingAdOwned"

type Interface interface {
	CreateDocument(ctx context.Context, from, to string) (string, error)
}

type BdAdOwned struct {
	Key  string `json:"_key"`
	From string `json:"_from"`
	To   string `json:"_to"`
}
