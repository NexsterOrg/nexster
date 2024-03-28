package social_graph

import (
	"context"

	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
	ytapi "github.com/NamalSanjaya/nexster/timeline/pkg/client/youtube_api"
	tp "github.com/NamalSanjaya/nexster/timeline/pkg/types"
)

type Interface interface {
	ListRecentPosts(ctx context.Context, userId, lastPostTimestamp, visibility string, noOfPosts int) ([]*map[string]interface{}, error)
	ListFriendSuggestions(ctx context.Context, userId string, offset, count int) ([]*map[string]string, error)
	UpdateMediaReaction(ctx context.Context, fromUserKey, toMediaKey, key string, newDoc map[string]interface{}) (string, error)
	ListOwnersPosts(ctx context.Context, userKey, lastPostTimestamp string, noOfPosts int) ([]*map[string]interface{}, error)
	CreateMediaReaction(ctx context.Context, fromUserKey, toMediaKey string, newDoc map[string]interface{}) (string, error)
	GetRole(authUserKey, userKey string) urepo.UserRole
	ListAllMedia(ctx context.Context, userKey string, offset, count int) ([]*map[string]string, error)
	ListPublicMedia(ctx context.Context, userKey string, offset, count int) ([]*map[string]string, error)
	GetUserKeyByIndexNo(ctx context.Context, indexNo string) (string, error)
	ListFriendSuggsV2(ctx context.Context, userKey, birthday, faculty, gender string, page, pageSize int) ([]*map[string]string, error)
	AttachFriendState(ctx context.Context, reqstorKey, friendKey string) (state string, reqId string, err error)
	CreateImagePost(ctx context.Context, userKey string, data *tp.Post) (string, string, error)
	DeleteImagePost(ctx context.Context, userKey, mediaKey string) error
	StoreVideosForFeed(ctx context.Context, ytClient *ytapi.YoutubeApi, interestCountPerUpdate int) error
}
