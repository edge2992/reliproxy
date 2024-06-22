package tests

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"third-party-proxy/internal/mocks"
	"third-party-proxy/pkg/handlers"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestHandleRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cbSettings := gobreaker.Settings{
		Name:        "HTTP GET",
		MaxRequests: 5,
		Interval:    2 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.TotalFailures > 3
		},
	}
	circuitBreaker := gobreaker.NewCircuitBreaker(cbSettings)

	t.Run("rate limit exceeded", func(t *testing.T) {
		rateLimiter := rate.NewLimiter(1, 1)
		mockClient := new(mocks.MockClient)
		handler := handlers.NewHandler(mockClient, circuitBreaker, rateLimiter, 1)

		router := gin.Default()
		router.GET("/proxy", handler.HandleRequest)

		assert.True(t, rateLimiter.Allow())
		assert.False(t, rateLimiter.Allow())

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/proxy", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.JSONEq(t, `{"error":"Rate limit exceeded"}`, w.Body.String())
	})

	t.Run("successful request", func(t *testing.T) {
		rateLimiter := rate.NewLimiter(1000, 1)
		mockClient := new(mocks.MockClient)
		handler := handlers.NewHandler(mockClient, circuitBreaker, rateLimiter, 1)

		router := gin.Default()
		router.GET("/proxy", handler.HandleRequest)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("Success")),
		}

		mockClient.On("Get", "https://api.thirdparty.com/data").Return(resp, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/proxy", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"data":"Success"}`, w.Body.String())
		mockClient.AssertExpectations(t)
	})
}
