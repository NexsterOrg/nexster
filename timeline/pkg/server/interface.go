package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// message body information
const ContentType string = "Content-Type"
const ContentLength string = "Content-Length"
const AllowOrigin string = "Access-Control-Allow-Origin"
const AllowMethods string = "Access-Control-Allow-Methods"

const AllMethods string = "GET, POST, PUT, DELETE, OPTIONS"

const Date string = "Date"

// message body information - ContentType
const ApplicationJson_Utf8 string = "application/json; charset=utf-8"

type Interface interface {
	ListRecentPostsForTimeline(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ListFriendSuggestions(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	UpdateMediaReactions(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ListPostsForOwnersTimeline(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	CreateMediaReactions(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	ListOwnersViewMedia(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	ListPublicMedia(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ListRoleBasedMedia(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ListFriendSuggestionsV2(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	CreateImagePost(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	DeleteImagePost(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type Reaction struct {
	MediaId    string `json:"media_id"`
	ReactorId  string `json:"reactor_id"`
	Key        string `json:"reaction_id"`
	Like       bool   `json:"like,omitempty"`
	Love       bool   `json:"love"`
	Laugh      bool   `json:"laugh"`
	Sad        bool   `json:"sad"`
	Insightful bool   `json:"insightful,omitempty"`
}
