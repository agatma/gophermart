package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func Encode(bytes []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(bytes)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
