package httpclient

import "net/http"

type DefaultHttpClient struct{}

func (c *DefaultHttpClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}
