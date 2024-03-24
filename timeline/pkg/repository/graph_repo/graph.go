package graphrepo

import (
	"context"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	ig "github.com/NamalSanjaya/nexster/pkgs/models/interestGroups"
	usr "github.com/NamalSanjaya/nexster/pkgs/models/user"
)

type repo struct {
	db argdb.Interface
}

var _ Interface = (*repo)(nil)

func NewRepo(argIntfce argdb.Interface) *repo {
	return &repo{db: argIntfce}
}

func (r *repo) ListInterestGroups(ctx context.Context, userKey string) (grps []*ig.InterestGroup, err error) {
	grps = []*ig.InterestGroup{}

	mpList, err := r.db.ListJsonAnyValue(ctx, listInterestGroupsQry, map[string]interface{}{
		"userNode": usr.MkUserDocId(userKey),
	})

	if err != nil {
		return
	}

	for _, doc := range mpList {
		k := (*doc)["key"].(string)
		if k == "" {
			continue
		}
		n := (*doc)["name"].(string)
		t := (*doc)["type"].(string)
		grps = append(grps, &ig.InterestGroup{
			Key:  k,
			Name: n,
			Type: t,
		})
	}

	return

}
