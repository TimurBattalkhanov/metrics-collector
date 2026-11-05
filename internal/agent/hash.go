package agent

import (
	"encoding/json"

	"github.com/TimurBattalkhanov/metrics-collector/internal/sign"
	"github.com/go-resty/resty/v2"
)

func addKey(key string) resty.RequestMiddleware {
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

		r.SetHeader("HashSHA256", sign.CalcHash(key, data))
		return nil
	}
}
