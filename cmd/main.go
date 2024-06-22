package main

import (
	"reliproxy/pkg/handlers"
	"reliproxy/pkg/httpclient"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

func main() {
	// サーキットブレーカーの設定
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

	rateLimiter := rate.NewLimiter(rate.Limit(5), 10)
	httpClient := &httpclient.DefaultHttpClient{}

	handler := handlers.NewHandler(httpClient, circuitBreaker, rateLimiter, 3)

	// Ginルーターの設定
	r := gin.Default()
	r.GET("/proxy", handler.HandleRequest)

	// サーバーの起動
	r.Run(":8080")
}
