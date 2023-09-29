package hmac

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"log"
)

func getHMAC(key string, values ...string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	pattern := ""
	for _, value := range values {
		pattern += value
	}
	h.Write([]byte(pattern))
	return h.Sum(nil)
}

func CalculateHMAC(key string, values ...string) string {
	return base64.URLEncoding.EncodeToString(getHMAC(key, values...))
}

func ValidateHMAC(key string, txHmacStr string, values ...string) bool {
	txHMACDecoded, err := base64.URLEncoding.DecodeString(txHmacStr)
	if err != nil {
		log.Println("failed to decode to base64 string: ", err)
		return false
	}
	rxHmac := getHMAC(key, values...)
	if hmac.Equal(txHMACDecoded, rxHmac) {
		return true
	} else {
		return false
	}
}
