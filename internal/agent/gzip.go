package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"

	"github.com/go-resty/resty/v2"
)

func gzipCompressBody(_ *resty.Client, r *resty.Request) error {
	if r.Body == nil {
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

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(data); err != nil {
		return err
	}
	if err := gzw.Close(); err != nil {
		return err
	}

	r.SetBody(buf.Bytes())
	r.SetHeader("Content-Encoding", "gzip")
	return nil
}
