package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Interface interface {
	SearchForUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
}
