package jwt

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const publicKeyPemFile string = "path-to-public_key.pem"
const privateKeyPemFile string = "path-to-private_key_pkcs8.pem"

const (
	tokenName    string = "token"
	loginPageUrl string = "http://192.168.1.101/test" // TODO: Need to change to Login page URL
	refTime      string = "2023-01-01T00:00:00.000Z"
)

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

func NewHandler(iss string, asAud string, h http.Handler) *JwtAuthHandler {
	return &JwtAuthHandler{
		issuer:     iss,
		asAudience: asAud,
		handler:    h,
	}
}

func (jah *JwtAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie(tokenName)
	// "token" is not found in cookies
	if err != nil {
		log.Println("failed to extract cookie: ", err)
		http.Redirect(w, r, loginPageUrl, http.StatusSeeOther)
		return
	}
	subject, err := jah.validateToken(tokenCookie.Value)
	// jwt token is invalid
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:  tokenName,
			Value: "",
			// Secure:   true, // TODO: Enable Secure: true, once you have the https connection.
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   -1,
		})
		log.Println("failed to validate cookie: ", err)
		http.Redirect(w, r, loginPageUrl, http.StatusSeeOther)
		return
	}
	// shallow copy of req with user_key
	jah.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), JwtUserKey, subject)))
}

// TODO:
// subject check and handover methods need to be added.
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
			"exp": time.Now().Add(12 * time.Minute).Unix(), // 12 Min valid time
		})
	signedToken, err := t.SignedString(key)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
