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
	PasswordResetLinkPath       string = "/usrmgmt/auth/password/reset-link"
	ForgotPasswordResetPath     string = "/usrmgmt/auth/password/reset"
)

var allowPaths []string = []string{
	AccessTokenPath, AccountCreationLinkPath, AccCreationLinkValidatePath,
	AccCreatePath, PasswordResetLinkPath, ForgotPasswordResetPath,
}

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
	// if strings.HasPrefix(urlPath, setTokenPath) {
	// 	ap.authCtrler.AuthDisabledServeHTTP(w, r)
	// 	return
	// }
	// if urlPath == AccessTokenPath || urlPath == AccountCreationLinkPath || urlPath == AccCreationLinkValidatePath ||
	// 	urlPath == AccCreatePath || urlPath == PasswordResetLinkPath {
	// 	ap.authCtrler.AuthDisabledServeHTTP(w, r)
	// 	return
	// }
	if isGivenIn(r.URL.Path, allowPaths) {
		ap.authCtrler.AuthDisabledServeHTTP(w, r)
		return
	}
	ap.authCtrler.ServeHTTP(w, r)
}

func isGivenIn(value string, values []string) bool {
	for _, each := range values {
		if each == value {
			return true
		}
	}
	return false
}
