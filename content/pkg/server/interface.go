package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// message body fields
const ContentType string = "Content-Type"
const Date string = "Date"

// message body values
const ApplicationJson_Utf8 string = "application/json; charset=utf-8"

type ServerConfig struct {
	SecretImgKey string `yaml:"secretImgKey"`
}

type Interface interface {
	ServeImages(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	CreateImgUrl(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}
