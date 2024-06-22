package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"reliproxy/pkg/httpclient"
	"reliproxy/pkg/utils"
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
				return nil, utils.ErrRateLimitExceeded
			}

			resp, err := h.client.Get("https://api.thirdparty.com/data")
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("%w: %d", utils.ErrUnexpectedStatusCode, resp.StatusCode)
			}

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			return string(bodyBytes), nil
		}, h.maxRetries)
	})

	if err != nil {
		switch {
		case errors.Is(err, utils.ErrRateLimitExceeded):
			utils.Logger.WithFields(logrus.Fields{
				"error": err,
			}).Warn("Rate limit exceeded")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		case errors.Is(err, utils.ErrUnexpectedStatusCode):
			utils.Logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("Unexpected status code")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected status code"})
		default:
			utils.Logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("Internal server error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}

		c.Error(err)
		return
	}

	utils.Logger.WithFields(logrus.Fields{
		"result": result,
	}).Info("Request handled successfully")
	c.JSON(http.StatusOK, gin.H{"data": result})
}
