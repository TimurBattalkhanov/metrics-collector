package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func CalcHash(key string, data []byte) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(data)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
