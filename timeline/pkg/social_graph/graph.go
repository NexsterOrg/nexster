package social_graph

import (
	"context"

	mrepo "github.com/NamalSanjaya/nexster/timeline/pkg/repos/media"
	rrepo "github.com/NamalSanjaya/nexster/timeline/pkg/repos/reaction"
	urepo "github.com/NamalSanjaya/nexster/timeline/pkg/repos/user"
)

// TODO
// 1. Change collection names, field names and other parameter names (eg: userRelations, mediaOwnerEdges)

const recentMediaQuery string = `FOR v,e,p IN 1..2 INBOUND @userNode userRelations, mediaOwnerEdges
	FILTER e.kind == "media_owner" && v.visibility == @visibility
	&& v.created_date <= DATE_ISO8601(@lastPostAt)
	SORT v.created_date DESC
	LIMIT @noOfPosts
	RETURN DISTINCT {"link" : v.link, "title" : v.title, 
	"description" : v.description,"created_date" : v.created_date, "size" : v.size }`

const suggestFriendsQuery string = `FOR v,e,p IN 2..2 OUTBOUND
	@userNode userRelations
	OPTIONS { uniqueVertices: "path" }
	FILTER e.kind == "friend" 
	&& e.started_at > DATE_ISO8601(@startedThreshold)
	SORT e.started_at
	LIMIT @noOfSuggestions
	RETURN { "user_id" : v.user_id, "username" : v.username, "image_url": v.image_url }`

type socialGraph struct {
	mediaRepo mrepo.Interface
	userRepo  urepo.Interface
	reactRepo rrepo.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewRepo(mIntfce mrepo.Interface, uIntfce urepo.Interface, rIntfce rrepo.Interface) *socialGraph {
	return &socialGraph{
		mediaRepo: mIntfce,
		userRepo:  uIntfce,
		reactRepo: rIntfce,
	}
}

func (sgr *socialGraph) ListRecentPosts(ctx context.Context, userId, lastPostTimestamp, visibility string, noOfPosts int) ([]*mrepo.Media, error) {
	bindVars := map[string]interface{}{
		"userNode":   sgr.userRepo.MkUserDocId(userId),
		"lastPostAt": lastPostTimestamp,
		"noOfPosts":  noOfPosts,
		"visibility": visibility,
	}
	return sgr.mediaRepo.ListMedia(ctx, recentMediaQuery, bindVars)
}

func (sgr *socialGraph) ListFriendSuggestions(ctx context.Context, userId, startedThreshold string, noOfSuggestions int) ([]*urepo.User, error) {
	bindVars := map[string]interface{}{
		"userNode":         sgr.userRepo.MkUserDocId(userId),
		"startedThreshold": startedThreshold,
		"noOfSuggestions":  noOfSuggestions,
	}
	return sgr.userRepo.ListUsers(ctx, suggestFriendsQuery, bindVars)
}

func (sgr *socialGraph) UpdateMediaReaction(ctx context.Context, fromUserKey, toMediaKey, key string, newDoc map[string]interface{}) error {
	return sgr.reactRepo.UpdateReactions(ctx, sgr.userRepo.MkUserDocId(fromUserKey), sgr.mediaRepo.MkMediaDocId(toMediaKey), key, newDoc)
}
