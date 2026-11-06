package agent

import (
	"time"

	"github.com/go-resty/resty/v2"
)

func retryAfter(retryDelays []time.Duration) resty.RetryAfterFunc {
	return func(_ *resty.Client, resp *resty.Response) (time.Duration, error) {
		return retryDelays[resp.Request.Attempt-1], nil
	}
}

func NewClient(key string, retryDelays []time.Duration) *resty.Client {
	c := resty.New().
		SetTimeout(time.Second).
		SetRetryCount(len(retryDelays)).
		SetRetryMaxWaitTime(retryDelays[len(retryDelays)-1]).
		SetRetryAfter(retryAfter(retryDelays))
	if len(key) > 0 {
		c.OnBeforeRequest(addKey(key))
	}
	c.OnBeforeRequest(gzipCompressBody)
	return c
}
