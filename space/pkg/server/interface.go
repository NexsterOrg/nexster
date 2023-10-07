package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Interface interface {
	CreateEventInSpace(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}
