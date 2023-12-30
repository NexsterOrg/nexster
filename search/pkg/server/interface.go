package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ServerConfig struct {
	AllowedOrigins []string `yaml:"allowedOrigins"`
}

type Interface interface {
	SearchForUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
}
