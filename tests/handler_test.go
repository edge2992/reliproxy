package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"third-party-proxy/internal/mocks"
	"third-party-proxy/pkg/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestHandleRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	client := new(mocks.MockClient)
	httpGet = client.Get

	router.GET("/proxy", handlers.HandleRequest)

	t.Run("rate limit exceeded", func(t *testing.T) {
		rateLimiter = rate.NewLimiter(1, 1)
		rateLimiter.Allow()
		rateLimiter.Allow()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/proxy", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.JSONEq(t, `{"error":"Rate limit exceeded"}`, w.Body.String())

	})
}
