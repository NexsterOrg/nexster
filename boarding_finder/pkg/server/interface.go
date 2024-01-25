package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Interface interface {
	CreateAd(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	CreateBoardingOwner(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	GetAdForMainView(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ChangeStatusOfAd(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type ServerConfig struct {
	ProjectDir     string   `yaml:"projectDir"`
	PublicKeyPath  string   `yaml:"publicKeyPath"`
	AllowedOrigins []string `yaml:"allowedOrigins"`
}
