package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func Sign(secret, value string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(value))
	return value + "." + hex.EncodeToString(mac.Sum(nil))
}

func Verify(secret, signedValue string) (string, bool) {
	value, sig, found := strings.Cut(signedValue, ".")
	if !found {
		return "", false
	}
	expected, err := hex.DecodeString(sig)
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(value))
	if !hmac.Equal(expected, mac.Sum(nil)) {
		return "", false
	}
	return value, true
}
