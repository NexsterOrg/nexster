package socialgraph

import (
	"context"

	dtm "github.com/NamalSanjaya/nexster/boarding_finder/pkg/dto_mapper"
)

type Interface interface {
	CreateAd(ctx context.Context, bdOwnerKey string, data *dtm.CreateAdDto) (adNodeKey, ownedEdgeKey string, err error)
	CreateBoardingOwner(ctx context.Context, data *dtm.CreateBoardingOwner) (bdOwnerKey string, err error)
	GetAdForMainView(ctx context.Context, adKey string) (adWithOwner *dtm.AdsWithOwner, err error)
	ChangeAdStatus(ctx context.Context, adKey, status string) error
	ListAdsWithFilters(ctx context.Context, data *dtm.ListFilterQueryParams) (ads []*dtm.AdForList, adsCount int, err error)
	IsBoardingOwnerExist(ctx context.Context, phoneNo string) (bool, error)
}
