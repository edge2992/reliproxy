package services

import (
	"net/http"
)

type IHTTPClient interface {
	Get(url string) (*http.Response, error)
}
