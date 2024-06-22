package httpclient

import (
	"fmt"
	"net/http"
	"reliproxy/pkg/utils"

	"github.com/sony/gobreaker"

	"golang.org/x/time/rate"
)

type ReliClient struct {
	client         HttpClient
	circuitBreaker *gobreaker.CircuitBreaker
	rateLimiter    *rate.Limiter
	maxRetries     int
}

func NewReliClient(client HttpClient, circuitBreaker *gobreaker.CircuitBreaker, rateLimiter *rate.Limiter, maxRetries int) *ReliClient {
	return &ReliClient{
		client:         client,
		circuitBreaker: circuitBreaker,
		rateLimiter:    rateLimiter,
		maxRetries:     maxRetries,
	}
}

func (r *ReliClient) Get(url string) (*http.Response, error) {
	var response *http.Response
	_, err := r.circuitBreaker.Execute(func() (interface{}, error) {
		return utils.RetryWithExponentialBackoff(func() (*http.Response, error) {
			if !r.rateLimiter.Allow() {
				return nil, utils.ErrRateLimitExceeded
			}

			resp, err := r.client.Get(url)
			if err != nil {
				return nil, err
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return nil, fmt.Errorf("%w: %d", utils.ErrUnexpectedStatusCode, resp.StatusCode)
			}

			response = resp
			return response, nil
		}, r.maxRetries)
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}
