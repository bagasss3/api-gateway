package gateway

import (
	"net/http"
	"net/url"
	"time"
)

func probe(u *url.URL) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	reqURL := *u
	reqURL.Path = singleJoin(reqURL.Path, "/health")
	resp, err := client.Get(reqURL.String())
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
