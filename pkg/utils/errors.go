package utils

import "errors"

var (
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
)
