package authprovider

import (
	"net/http"
	"strings"

	jwtAuth "github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
)

const (
	AccessTokenPath             string = "/usrmgmt/auth/token"
	AccountCreationLinkPath     string = "/usrmgmt/auth/reg-link"
	AccCreationLinkValidatePath string = "/usrmgmt/auth/reg-link/validate"
	registerPath                string = "/usrmgmt/auth/reg"
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
	// TODO: temporty path to get token for development work
	if strings.HasPrefix(r.URL.Path, setTokenPath) {
		ap.authCtrler.AuthDisabledServeHTTP(w, r)
		return
	}
	if r.URL.Path == AccessTokenPath || r.URL.Path == AccountCreationLinkPath ||
		r.URL.Path == AccCreationLinkValidatePath {
		ap.authCtrler.AuthDisabledServeHTTP(w, r)
		return
	}
	ap.authCtrler.ServeHTTP(w, r)
}
