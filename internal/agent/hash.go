package agent

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/go-resty/resty/v2"
)

func addKeyIfExist(key string) resty.RequestMiddleware {
	if key == "" {
		return func(_ *resty.Client, _ *resty.Request) error {
			return nil
		}
	}
	return func(_ *resty.Client, r *resty.Request) error {
		if r.Body == nil || r.Header.Get("HashSHA256") != "" {
			return nil
		}

		data, ok := r.Body.([]byte)
		if !ok {
			var err error
			data, err = json.Marshal(r.Body)
			if err != nil {
				return err
			}
		}

		h := hmac.New(sha256.New, []byte(key))
		h.Write(data)
		dst := h.Sum(nil)

		r.SetHeader("HashSHA256", hex.EncodeToString(dst))
		return nil
	}
}
