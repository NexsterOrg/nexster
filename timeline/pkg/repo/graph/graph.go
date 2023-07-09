package graph

import (
	"context"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

// TODO
// 1. Change collection names, field names and other parameter names

const recentMediaQuery string = `FOR v,e,p IN 1..2 INBOUND @userNode userRelations, mediaOwnerEdges
	FILTER e.kind == "media_owner" && v.visibility == "public"
	&& v.created_date <= DATE_ISO8601(@lastPostAt)
	SORT v.created_date DESC
	LIMIT @noOfPosts
	RETURN DISTINCT {"link" : v.link, "title" : v.title, 
	"description" : v.description,"created_date" : v.created_date, "size" : v.size }`

type graphRepo struct {
	client argdb.Interface
}

var _ Interface = (*graphRepo)(nil)

func NewRepo(argdbInterface argdb.Interface) *graphRepo {
	return &graphRepo{
		client: argdbInterface,
	}
}

func (gr *graphRepo) GetPostsForTimeline(ctx context.Context, userNode, lastPostTimestamp string, noOfPosts int) ([]*argdb.Media, error) {
	bindVars := map[string]interface{}{
		"userNode":   userNode,
		"lastPostAt": lastPostTimestamp,
		"noOfPosts":  noOfPosts,
	}
	return gr.client.ListMedia(ctx, recentMediaQuery, bindVars)
}
