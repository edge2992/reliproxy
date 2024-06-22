package services

import "net/http"

type HTTPClient struct{}

func (c *HTTPClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}
