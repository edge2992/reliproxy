package main

import (
	"reliproxy/pkg/db"
	"reliproxy/pkg/handlers"
	"reliproxy/pkg/httpclient"
	"reliproxy/pkg/models"
	"reliproxy/pkg/queue"
	"reliproxy/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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

	connectionEnv := db.NewMySQLConnectionEnv()
	dbn, err := connectionEnv.ConnectDBWithRetry()
	if err != nil {
		panic(err)
	}
	db.Migrate(dbn)
	statusRepository := models.NewGormRequestStatusRepository(dbn)

	rdb := redis.NewClient(&redis.Options{
		Addr:     utils.GetEnv("REDIS_ADDR", "localhost:6379"),
		Password: utils.GetEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})
	queue := queue.NewRedisQueue(rdb, utils.GetEnv("REDIS_QUEUE_NAME", "queue"))

	asyncWriteHandler := handlers.NewAsyncWriteHandler(queue, statusRepository)

	// Ginルーターの設定
	r := gin.Default()
	r.GET("/proxy", handler.HandleRequest)
	r.GET("/async-proxy", asyncWriteHandler.HandleRequest)

	// サーバーの起動
	r.Run(":8080")
}
