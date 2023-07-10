package social_graph

import (
	"context"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
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
	graph argdb.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewRepo(argdbInterface argdb.Interface) *socialGraph {
	return &socialGraph{
		graph: argdbInterface,
	}
}

func (sgr *socialGraph) ListRecentPosts(ctx context.Context, userNode, lastPostTimestamp, visibility string, noOfPosts int) (Posts, error) {
	bindVars := map[string]interface{}{
		"userNode":   userNode,
		"lastPostAt": lastPostTimestamp,
		"noOfPosts":  noOfPosts,
		"visibility": visibility,
	}
	return sgr.graph.ListMedia(ctx, recentMediaQuery, bindVars)
}

func (sgr *socialGraph) ListFriendSuggestions(ctx context.Context, userNode, startedThreshold string, noOfSuggestions int) (Users, error) {
	bindVars := map[string]interface{}{
		"userNode":         userNode,
		"startedThreshold": startedThreshold,
		"noOfSuggestions":  noOfSuggestions,
	}
	return sgr.graph.ListUsers(ctx, suggestFriendsQuery, bindVars)
}
