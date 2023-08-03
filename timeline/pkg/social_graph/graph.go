package social_graph

import (
	"context"
	"fmt"
	"log"

	mrepo "github.com/NamalSanjaya/nexster/pkgs/models/media"
	rrepo "github.com/NamalSanjaya/nexster/pkgs/models/reaction"
	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
)

// TODO
// 1. Change collection names, field names and other parameter names (eg: friends, mediaOwnerEdges)

// TODO: For users/482201 case return wrong results.
const recentMediaQuery string = `FOR v,e IN 1..2 INBOUND @userNode friends, mediaOwnerEdges
	FILTER e.kind == "media_owner" && v.visibility == @visibility
	&& v.created_date < DATE_ISO8601(@lastPostAt)
	SORT v.created_date DESC
	LIMIT @noOfPosts
	RETURN DISTINCT {"media": {"_key": v._key, "link" : v.link, "title" : v.title, 
	"description" : v.description,"created_date" : v.created_date, "size" : v.size}, "owner_id": e._to}`

const suggestFriendsQuery string = `FOR v,e IN 2..2 OUTBOUND
	@userNode friends
	OPTIONS { uniqueVertices: "path" }
	FILTER e.kind == "friend" 
	&& e.started_at > DATE_ISO8601(@startedThreshold)
	COLLECT key = v._key INTO groups
	SORT null
	SORT groups[0].e.started_at
	LIMIT @noOfSuggestions
	RETURN {"key" : key, "username" : groups[0].v.username, "image_url": groups[0].v.image_url, "faculty": groups[0].v.faculty, 
	"field": groups[0].v.degree_info.field, "batch": groups[0].v.batch, "friendship_started": groups[0].e.started_at }`

const getOrder1FriendsQuery string = `FOR v,e IN 1..1 OUTBOUND
	@userNode friends
	FILTER e.kind == "friend"
	RETURN v._key`

const getReactionQuery string = `FOR v,e IN 1..1 INBOUND @mediaNode reactions
    RETURN { "like": e["like"], "love": e.love, "laugh": e.laugh,
    "sad": e.sad, "insightful": e.insightful }`

const getOwnersMediaQuery string = `FOR v,e IN 1..1 INBOUND @userNode mediaOwnerEdges
	FILTER v.created_date <= DATE_ISO8601(@lastPostAt)
	SORT v.created_date DESC
	LIMIT @noOfPosts
	RETURN DISTINCT {"_key": v._key, "link" : v.link, "title" : v.title, 
	"description" : v.description,"created_date" : v.created_date, "size" : v.size}`

const getViewerReactions string = `FOR r IN reactions
	FILTER r._from == @fromUser AND r._to == @toMedia
	LIMIT 1
	RETURN r`

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

		viewersReacts, err2 := sgr.reactRepo.GetViewersReactions(ctx, getViewerReactions, map[string]interface{}{
			"fromUser": sgr.userRepo.MkUserDocId(userId), "toMedia": sgr.mediaRepo.MkMediaDocId(media.Media.Key),
		})
		if err2 != nil {
			log.Println(err2)
			continue
		}

		posts = append(posts, &map[string]interface{}{
			"media": media.Media, "owner": map[string]string{"_key": user.UserId, "name": user.Username, "Headling": user.Headling, "image_url": user.ImageUrl},
			"reactions": racts, "viewer_reaction": map[string]interface{}{"key": viewersReacts.Key, "like": viewersReacts.Like, "love": viewersReacts.Love,
				"laugh": viewersReacts.Laugh},
		})
	}

	return posts, nil
}

func (sgr *socialGraph) ListOwnersPosts(ctx context.Context, userKey, lastPostTimestamp string, noOfPosts int) ([]*map[string]interface{}, error) {
	posts := []*map[string]interface{}{}
	bindVars := map[string]interface{}{
		"userNode":   sgr.userRepo.MkUserDocId(userKey),
		"lastPostAt": lastPostTimestamp,
		"noOfPosts":  noOfPosts,
	}
	medias, err := sgr.mediaRepo.ListMedia(ctx, getOwnersMediaQuery, bindVars)
	if err != nil {
		return posts, err
	}

	for _, media := range medias {
		racts, err2 := sgr.reactRepo.GetReactionsCount(ctx, getReactionQuery, map[string]interface{}{
			"mediaNode": sgr.mediaRepo.MkMediaDocId(media.Key),
		})
		if err2 != nil {
			log.Println(err2)
			continue
		}
		posts = append(posts, &map[string]interface{}{
			"media":     media,
			"reactions": racts,
		})
	}

	return posts, nil
}

func (sgr *socialGraph) ListFriendSuggestions(ctx context.Context, userId, startedThreshold string, noOfSuggestions int) ([]*map[string]string, error) {
	userDocId := sgr.userRepo.MkUserDocId(userId)
	bindVars1 := map[string]interface{}{
		"userNode":         userDocId,
		"startedThreshold": startedThreshold,
		"noOfSuggestions":  noOfSuggestions,
	}
	order2Nodes, err := sgr.userRepo.ListUsersV2(ctx, suggestFriendsQuery, bindVars1)
	if err != nil {
		return order2Nodes, fmt.Errorf("falied to list 2nd order user info due to %v", err)
	}
	order1Friends, err := sgr.userRepo.ListStrings(ctx, getOrder1FriendsQuery, map[string]interface{}{"userNode": userDocId})
	if err != nil {
		return order2Nodes, fmt.Errorf("failed to list 1st order users due to %v", err)
	}
	results := []*map[string]string{}
	// remove 1st order nodes from 2nd order nodes
	for _, node2 := range order2Nodes {
		notFound := true
		for _, key1 := range order1Friends {
			if *key1 == (*node2)["key"] {
				notFound = false
				break
			}
		}
		if notFound {
			results = append(results, node2)
		}
	}
	return results, nil
}

func (sgr *socialGraph) UpdateMediaReaction(ctx context.Context, fromUserKey, toMediaKey, key string, newDoc map[string]interface{}) (string, error) {
	return sgr.reactRepo.UpdateReactions(ctx, sgr.userRepo.MkUserDocId(fromUserKey), sgr.mediaRepo.MkMediaDocId(toMediaKey), key, newDoc)
}

func (sgr *socialGraph) CreateMediaReaction(ctx context.Context, fromUserKey, toMediaKey string, newDoc map[string]interface{}) (string, error) {
	// first check whether a there is a link or not
	viewersReacts, err := sgr.reactRepo.GetViewersReactions(ctx, getViewerReactions, map[string]interface{}{
		"fromUser": sgr.userRepo.MkUserDocId(fromUserKey), "toMedia": sgr.mediaRepo.MkMediaDocId(toMediaKey),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create reaction link due to %v", err)
	}
	// Already there is a key
	if viewersReacts.Key != "" {
		return sgr.reactRepo.UpdateReactions(ctx, sgr.userRepo.MkUserDocId(fromUserKey), sgr.mediaRepo.MkMediaDocId(toMediaKey), viewersReacts.Key, newDoc)
	}
	return sgr.reactRepo.CreateReactionLink(ctx, sgr.userRepo.MkUserDocId(fromUserKey), sgr.mediaRepo.MkMediaDocId(toMediaKey), newDoc)
}
