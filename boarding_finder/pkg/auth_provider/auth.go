package authprovider

import (
	"net/http"

	jwtAuth "github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
)

// /bdfinder/auth/* ---> auth disabled paths
const (
	BdOwnerAccCreatePath string = "/bdfinder/auth/owner"
	SmsOtpSendPath       string = "/bdfinder/otp/send"
)

type authProvider struct {
	authCtrler jwtAuth.Interface
}

func NewAuthProvider(jwtIntfce jwtAuth.Interface) *authProvider {
	return &authProvider{
		authCtrler: jwtIntfce,
	}
}

func (ap *authProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path

	if urlPath == BdOwnerAccCreatePath {
		ap.authCtrler.AuthDisabledServeHTTP(w, r)
		return
	}
	ap.authCtrler.ServeHTTP(w, r)
}
