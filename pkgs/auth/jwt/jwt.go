package jwt

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"

	strg "github.com/NamalSanjaya/nexster/pkgs/utill/string"
)

type jwtUserKeyType string
type rolesType string

const JwtUserKey jwtUserKeyType = "user_key"
const roles rolesType = "roles"

// Responsibilities
// Extract the JWT token, Validate sign/algo, validate claims, direct request to signup/signin pages if validation failed.
// We use same public key across all services.
type JwtAuthHandler struct {
	handler    http.Handler
	issuer     string
	asAudience string
	publicKey  []byte
}

var _ Interface = (*JwtAuthHandler)(nil)

func NewHandler(iss string, asAud string, h http.Handler, publicKeyPath string) *JwtAuthHandler {
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Panicf("failed to read publicKey file: %v", err)
	}
	return &JwtAuthHandler{
		issuer:     iss,
		asAudience: asAud,
		handler:    h,
		publicKey:  publicKey,
	}
}

func (jah *JwtAuthHandler) AuthDisabledServeHTTP(w http.ResponseWriter, r *http.Request) {
	jah.handler.ServeHTTP(w, r)
}

func (jah *JwtAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		jah.handler.ServeHTTP(w, r)
		return
	}
	jwtToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if jwtToken == "" {
		log.Println("No Authorization header is provided") // TODO: remove in productiono environemtn
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	subject, roleList, err := jah.validateToken(jwtToken)
	// jwt token is invalid
	if err != nil {
		log.Println("failed to validate token: ", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	ctx1 := context.WithValue(r.Context(), JwtUserKey, subject)
	// shallow copy of req with user_key
	jah.handler.ServeHTTP(w, r.WithContext(context.WithValue(ctx1, roles, roleList)))
}

func (jah *JwtAuthHandler) validateToken(tokenString string) (userKey string, roleList []string, err error) {
	// TODO: Need to think about how to read and share same public key acorss components
	roleList = []string{}
	publicKey, err := jwt.ParseECPublicKeyFromPEM(jah.publicKey)
	if err != nil {
		err = fmt.Errorf("failed parse EC public key: %v", err)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unsupport signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	// Can be due to expiration of JWT
	if err != nil {
		err = fmt.Errorf("failed to parse: %v", err)
		return
	}

	// Verify the token's signature
	if !token.Valid {
		err = fmt.Errorf("invalid token signature")
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = fmt.Errorf("invalid token claims")
		return
	}

	// We only support jwt issued by usrmgmt server
	if claims["iss"].(string) != jah.issuer {
		err = fmt.Errorf("unsupport issuer %s", claims["iss"].(string))
		return
	}
	notValidAud := true
	if audClaims, ok := claims["aud"].([]interface{}); ok {
		for _, audClaim := range audClaims {
			if audStr, ok := audClaim.(string); ok {
				if audStr == jah.asAudience {
					notValidAud = false
					break
				}
			}
		}
	} else {
		err = fmt.Errorf("unsupport claims found in jwt payload")
		return
	}

	if notValidAud {
		err = fmt.Errorf("aud validation is falied")
		return
	}

	if userKey, ok = claims["sub"].(string); !ok {
		err = fmt.Errorf("failed to extract sub")
		return
	}
	roleInterfaceList := []interface{}{}

	if roleInterfaceList, ok = claims["roles"].([]interface{}); !ok {
		err = fmt.Errorf("failed to extract roles")
		return
	}
	roleList, err = strg.InterfaceToStringArray(roleInterfaceList)
	return
}

func GetUserKey(ctx context.Context) (string, error) {
	jwtUserKey, ok := ctx.Value(JwtUserKey).(string)
	if !ok {
		return "", fmt.Errorf("failed to extract user key")
	}
	return jwtUserKey, nil
}

func GetRoles(ctx context.Context) ([]string, error) {
	roleArr, ok := ctx.Value(roles).([]string)
	if !ok {
		return []string{}, fmt.Errorf("failed to extract roles")
	}
	return roleArr, nil
}
