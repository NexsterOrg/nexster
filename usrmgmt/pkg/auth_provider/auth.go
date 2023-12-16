package authprovider

import (
	"net/http"

	jwtAuth "github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
)

const (
	AccessTokenPath             string = "/usrmgmt/auth/token"
	AccountCreationLinkPath     string = "/usrmgmt/auth/reg-link"
	AccCreationLinkValidatePath string = "/usrmgmt/auth/reg-link/validate"
	AccCreatePath               string = "/usrmgmt/auth/reg"
	setTokenPath                string = "/usrmgmt/set-token/"
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

	// TODO: temporty path to get token for development work
	// if strings.HasPrefix(urlPath, setTokenPath) {
	// 	ap.authCtrler.AuthDisabledServeHTTP(w, r)
	// 	return
	// }
	if urlPath == AccessTokenPath || urlPath == AccountCreationLinkPath || urlPath == AccCreationLinkValidatePath ||
		urlPath == AccCreatePath {
		ap.authCtrler.AuthDisabledServeHTTP(w, r)
		return
	}
	ap.authCtrler.ServeHTTP(w, r)
}
