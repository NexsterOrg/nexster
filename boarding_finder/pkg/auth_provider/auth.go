package authprovider

import (
	"net/http"

	jwtAuth "github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
)

// /bdfinder/auth/* ---> auth disabled paths
const (
	BdOwnerAccCreatePath string = "/bdfinder/auth/owner"
	SmsOtpSendPath       string = "/bdfinder/otp/send"
	SmsOtpVerifyPath     string = "/bdfinder/otp/verify"
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

	if urlPath == BdOwnerAccCreatePath || urlPath == SmsOtpSendPath || urlPath == SmsOtpVerifyPath {
		ap.authCtrler.AuthDisabledServeHTTP(w, r)
		return
	}
	ap.authCtrler.ServeHTTP(w, r)
}
