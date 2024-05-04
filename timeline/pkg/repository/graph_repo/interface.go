package graphrepo

import (
	"context"

	ig "github.com/NamalSanjaya/nexster/pkgs/models/interestGroups"
)

type Interface interface {
	ListInterestGroups(ctx context.Context, userKey string) (grps []*ig.InterestGroup, err error)
}
