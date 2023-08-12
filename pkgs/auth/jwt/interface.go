package jwt

import (
	"net/http"
)

type Interface interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	AuthDisabledServeHTTP(w http.ResponseWriter, r *http.Request)
}
