package socialgraph

import (
	"context"

	dtm "github.com/NamalSanjaya/nexster/boarding_finder/pkg/dto_mapper"
)

type Interface interface {
	CreateAd(ctx context.Context, bdOwnerKey string, data *dtm.CreateAdDto) (adNodeKey, ownedEdgeKey string, err error)
	CreateBoardingOwner(ctx context.Context, data *dtm.CreateBoardingOwner) (bdOwnerKey string, err error)
}
