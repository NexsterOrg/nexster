package jwt

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const publicKeyPemFile string = "path-to-public_key.pem"
const privateKeyPemFile string = "path-to-private_key_pkcs8.pem"

type jwtUserKeyType string

const JwtUserKey jwtUserKeyType = "user_key"

// Responsibilities
// Extract the JWT token, Validate sign/algo, validate claims, direct request to signup/signin pages if validation failed.
// We use same public key across all services.
type JwtAuthHandler struct {
	handler    http.Handler
	issuer     string
	asAudience string
}

var _ Interface = (*JwtAuthHandler)(nil)

func NewHandler(iss string, asAud string, h http.Handler) *JwtAuthHandler {
	return &JwtAuthHandler{
		issuer:     iss,
		asAudience: asAud,
		handler:    h,
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
	// "token" is not found in cookies
	if jwtToken == "" {
		log.Println("No Authorization header is provided") // TODO: remove in productiono environemtn
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	subject, err := jah.validateToken(jwtToken)
	// jwt token is invalid
	if err != nil {
		log.Println("failed to validate token: ", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// shallow copy of req with user_key
	jah.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), JwtUserKey, subject)))
}

func (jah *JwtAuthHandler) validateToken(tokenString string) (string, error) {
	// TODO: Need to think about how to read and share same public key acorss components
	key, err := os.ReadFile(publicKeyPemFile)
	if err != nil {
		return "", fmt.Errorf("failed to read public key pem file: %v", err)
	}

	publicKey, err := jwt.ParseECPublicKeyFromPEM(key)
	if err != nil {
		return "", fmt.Errorf("failed parse EC public key: %v", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unsupport signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	// Can be due to expiration of JWT
	if err != nil {
		return "", fmt.Errorf("failed to parse: %v", err)
	}

	// Verify the token's signature
	if !token.Valid {
		return "", fmt.Errorf("invalid token signature")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// We only support jwt issued by usrmgmt server
	if claims["iss"].(string) != jah.issuer {
		return "", fmt.Errorf("unsupport issuer %s", claims["iss"].(string))
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
		return "", fmt.Errorf("unsupport claims found in jwt payload")
	}

	if notValidAud {
		return "", fmt.Errorf("aud validation is falied")
	}

	return claims["sub"].(string), nil
}

// This is not a capability of Auth Handler. This is only used by usrmgmt server to generate jwt token.
func GenJwtToken(issuer, subject string, audience []string) (string, error) {
	privateKey, err := os.ReadFile(privateKeyPemFile)
	if err != nil {
		return "", err
	}

	key, err := jwt.ParseECPrivateKeyFromPEM(privateKey)
	if err != nil {
		return "", err
	}

	t := jwt.NewWithClaims(jwt.SigningMethodES256,
		jwt.MapClaims{
			"iss": issuer,
			"sub": subject,
			"aud": audience,
			"exp": time.Now().Add(12 * time.Hour).Unix(), // 12hr valid time
		})
	signedToken, err := t.SignedString(key)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
