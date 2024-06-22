package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"third-party-proxy/pkg/utils"
)

var (
	rdb            *redis.Client
	ctx            = context.Background()
	cb             *gobreaker.CircuitBreaker
	rateLimiter    *rate.Limiter
	requestPerSec  = 5
	burstAllowance = 10
	maxRetries     = 3
)

func init() {
	// Redisクライアントの初期化
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// サーキットブレーカーの設定
	settings := gobreaker.Settings{
		Name:        "HTTP GET",
		MaxRequests: 5,
		Interval:    2 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.TotalFailures > 3
		},
	}
	cb = gobreaker.NewCircuitBreaker(settings)

	// レートリミッターの設定
	rateLimiter = rate.NewLimiter(rate.Limit(requestPerSec), burstAllowance)
}

func HandleRequest(c *gin.Context) {
	result, err := cb.Execute(func() (interface{}, error) {
		return utils.RetryWithExponentialBackoff(func() (interface{}, error) {
			if !rateLimiter.Allow() {
				return nil, fmt.Errorf("rate limit exceeded")
			}

			resp, err := http.Get("https://api.thirdparty.com/data")
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}
			return "Success", nil
		}, maxRetries)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}
