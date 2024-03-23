package interestarray

import (
	"context"

	typs "github.com/NamalSanjaya/nexster/timeline/pkg/types"
)

type Interface interface {
	ListVideoIdsForFeed(ctx context.Context, userKey string, curPage, offset, limit int) (videos []*typs.StemVideoResp, count, nextPg int, err error)
}
