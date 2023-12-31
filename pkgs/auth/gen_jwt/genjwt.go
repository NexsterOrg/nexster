package genjwt

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type JwtTokenGenerator struct {
	issuer     string
	privateKey []byte
}

var _ Interface = (*JwtTokenGenerator)(nil)

func NewGenerator(issuer, privateKeyPath string) *JwtTokenGenerator {
	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Panicf("failed to read privateKey file: %v", err)
	}

	return &JwtTokenGenerator{
		issuer:     issuer,
		privateKey: privateKey,
	}
}

func (jtg *JwtTokenGenerator) GenJwtToken(subject string, audience []string) (string, error) {
	key, err := jwt.ParseECPrivateKeyFromPEM(jtg.privateKey)
	if err != nil {
		return "", err
	}

	t := jwt.NewWithClaims(jwt.SigningMethodES256,
		jwt.MapClaims{
			"iss": jtg.issuer,
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
