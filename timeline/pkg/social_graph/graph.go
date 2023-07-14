package social_graph

import (
	"context"
	"log"

	mrepo "github.com/NamalSanjaya/nexster/pkgs/models/media"
	rrepo "github.com/NamalSanjaya/nexster/pkgs/models/reaction"
	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
)

// TODO
// 1. Change collection names, field names and other parameter names (eg: friends, mediaOwnerEdges)

const recentMediaQuery string = `FOR v,e,p IN 1..2 INBOUND @userNode friends, mediaOwnerEdges
	FILTER e.kind == "media_owner" && v.visibility == @visibility
	&& v.created_date <= DATE_ISO8601(@lastPostAt)
	SORT v.created_date DESC
	LIMIT @noOfPosts
	RETURN DISTINCT {"media": {"_key": v._key, "link" : v.link, "title" : v.title, 
	"description" : v.description,"created_date" : v.created_date, "size" : v.size}, "owner_id": e._to}`

const suggestFriendsQuery string = `FOR v,e,p IN 2..2 OUTBOUND
	@userNode friends
	OPTIONS { uniqueVertices: "path" }
	FILTER e.kind == "friend" 
	&& e.started_at > DATE_ISO8601(@startedThreshold)
	SORT e.started_at
	LIMIT @noOfSuggestions
	RETURN { "user_id" : v.user_id, "username" : v.username, "image_url": v.image_url }`

const getReactionQuery string = `FOR v,e IN 1..1 INBOUND @mediaNode reactions
RETURN { "like": e["like"], "love": e.love, "laugh": e.laugh,
    "sad": e.sad, "insightful": e.insightful }`

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

func (sgr *socialGraph) ListRecentPosts(ctx context.Context, userId, lastPostTimestamp, visibility string, noOfPosts int) ([]*map[string]interface{}, error) {
	posts := []*map[string]interface{}{}
	bindVars := map[string]interface{}{
		"userNode":   sgr.userRepo.MkUserDocId(userId),
		"lastPostAt": lastPostTimestamp,
		"noOfPosts":  noOfPosts,
		"visibility": visibility,
	}
	medias, err := sgr.mediaRepo.ListMediaWithOwner(ctx, recentMediaQuery, bindVars)
	if err != nil {
		return posts, err
	}
	prefixLn := len(urepo.UsersColl) + 1 // length of "users/"

	for _, media := range medias {
		user, err2 := sgr.userRepo.GetUser(ctx, media.OwnerId[prefixLn:])
		if err2 != nil {
			log.Println(err2)
			continue
		}

		racts, err2 := sgr.reactRepo.GetReactionsCount(ctx, getReactionQuery, map[string]interface{}{
			"mediaNode": sgr.mediaRepo.MkMediaDocId(media.Media.Key),
		})
		if err2 != nil {
			log.Println(err2)
			continue
		}

		posts = append(posts, &map[string]interface{}{
			"media": media.Media, "owner": map[string]string{"_key": user.UserId, "name": user.Username, "Headling": user.Headling, "image_url": user.ImageUrl},
			"reactions": racts,
		})
	}

	return posts, nil
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
