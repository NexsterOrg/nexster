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
	RemovePendingFriendReq(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type FriendRequest struct {
	Key     string `json:"friendreq_id"`
	From    string `json:"requestor" validate:"required"`
	To      string `json:"friend" validate:"required"`
	Mode    string `json:"mode" validate:"required"`
	State   string `json:"state" validate:"required"`
	ReqDate string `json:"req_date" validate:"required"`
}
