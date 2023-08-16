package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// message body information
const ContentType string = "Content-Type"
const ContentLength string = "Content-Length"

const Date string = "Date"

// message body information - ContentType
const ApplicationJson_Utf8 string = "application/json; charset=utf-8"

type Interface interface {
	HandleFriendReq(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	RemovePendingFriendReq(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	RemoveFriendship(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	CreateFriendLink(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ListFriendInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetProfile(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetFriendsCount(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	SetCookie(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	SetAuthToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	GetUserKeyByIndexNo(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type FriendRequest struct {
	Key     string `json:"friendreq_id"`
	From    string `json:"requestor" validate:"required"`
	To      string `json:"friend" validate:"required"`
	Mode    string `json:"mode" validate:"required"`
	State   string `json:"state" validate:"required"`
	ReqDate string `json:"req_date" validate:"required"`
}

type FriendReqAcceptance struct {
	User1Key   string `json:"reqstor_id" validate:"required"`
	User2Key   string `json:"acceptor_id" validate:"required"`
	AcceptedAt string `json:"accepted_at" validate:"required"`
}
