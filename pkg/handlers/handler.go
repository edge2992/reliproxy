package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"third-party-proxy/pkg/httpclient"
	"third-party-proxy/pkg/utils"
)

type Handler struct {
	client         httpclient.HttpClient
	circuitBreaker *gobreaker.CircuitBreaker
	rateLimiter    *rate.Limiter
	maxRetries     int
}

func NewHandler(client httpclient.HttpClient, cb *gobreaker.CircuitBreaker, rl *rate.Limiter, maxRetries int) *Handler {
	return &Handler{
		client:         client,
		circuitBreaker: cb,
		rateLimiter:    rl,
		maxRetries:     maxRetries,
	}
}

func (h *Handler) HandleRequest(c *gin.Context) {
	result, err := h.circuitBreaker.Execute(func() (interface{}, error) {
		return utils.RetryWithExponentialBackoff(func() (interface{}, error) {
			if !h.rateLimiter.Allow() {
				return nil, fmt.Errorf("rate limit exceeded")
			}

			resp, err := h.client.Get("https://api.thirdparty.com/data")
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}
			return "Success", nil
		}, h.maxRetries)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}
