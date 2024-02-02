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
	ListAdsForMainView(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	SendOTP(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	VerifyOTP(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	DeleteAd(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ValidateUserForBdLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
}

type ServerConfig struct {
	ProjectDir     string   `yaml:"projectDir"`
	PublicKeyPath  string   `yaml:"publicKeyPath"`
	AllowedOrigins []string `yaml:"allowedOrigins"`
}

type OtpInfo struct {
	Otp      int
	ExpAt    int64
	Verified bool
}
