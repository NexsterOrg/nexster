package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Interface interface {
	ListRecentPostsForTimeline(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	ListFriendSuggestionsForTimeline(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
}

// message body information
const ContentType string = "Content-Type"
const ContentLength string = "Content-Length"

const Date string = "Date"

// message body information - ContentType
const ApplicationJson_Utf8 string = "application/json; charset=utf-8"
