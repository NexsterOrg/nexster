package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Interface interface {
	CreateEventInSpace(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ListUpcomingEventsFromSpace(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	GetEventFromSpace(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ListLoveReactUsersForEvent(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	CreateEventReaction(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	SetEventReactionState(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ListMyEventsFromSpace(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	DeleteEventFromSpace(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}
