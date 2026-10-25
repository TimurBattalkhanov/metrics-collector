package agent

import (
	"time"

	"github.com/go-resty/resty/v2"
)

var retryDelays = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

func retryAfter(_ *resty.Client, resp *resty.Response) (time.Duration, error) {
	return retryDelays[resp.Request.Attempt-1], nil
}

func getMaxWaitRetry() time.Duration {
	return retryDelays[len(retryDelays)-1]
}

func NewClient(key string) *resty.Client {
	return resty.New().
		SetTimeout(time.Second).
		SetRetryCount(len(retryDelays)).
		SetRetryMaxWaitTime(getMaxWaitRetry()).
		SetRetryAfter(retryAfter).
		OnBeforeRequest(addKeyIfExist(key)).
		OnBeforeRequest(gzipCompressBody)
}
