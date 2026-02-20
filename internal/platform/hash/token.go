package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func HashToken(token, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}
