package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

var (
	rdb            *redis.Client
	cb             *gobreaker.CircuitBreaker
	rateLimiter    *rate.Limiter
	requestPerSec  = 5
	burstAllowance = 10
	maxRetries     = 3
)

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379"})

	settings := gobreaker.Settings{
		Name:        "sample circuit breaker",
		MaxRequests: 5,
		Interval:    2 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.TotalFailures > 3
		},
	}

	cb = gobreaker.NewCircuitBreaker(settings)
	rateLimiter = rate.NewLimiter(rate.Limit(requestPerSec), burstAllowance)

	r := gin.Default()

	r.GET("/proxy", handleRequest)
	r.Run(":8080")
}

func handleRequest(c *gin.Context) {
	if !rateLimiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		return
	}

	result, err := cb.Execute(func() (interface{}, error) {
		return retryWithExponentialBackoff(func() (interface{}, error) {
			if !rateLimiter.Allow() {
				return nil, fmt.Errorf("Rate limit exceeded")
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
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}

// retryWithExponentialBackoff retries the operation with exponential backoff
func retryWithExponentialBackoff(operation func() (interface{}, error), maxRetries int) (interface{}, error) {
	var result interface{}
	var err error

	for i := 0; i < maxRetries; i++ {
		result, err = operation()
		if err == nil {
			return result, nil
		}
		backoffDuration := time.Duration((1 << i)) * time.Second
		time.Sleep(backoffDuration)
	}
	return nil, fmt.Errorf("operation failed after %d retries retries: %v", maxRetries, err)
}
