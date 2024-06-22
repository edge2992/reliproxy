package main

import (
	"reliproxy/pkg/consumer"
	"reliproxy/pkg/db"
	"reliproxy/pkg/handlers"
	"reliproxy/pkg/httpclient"
	"reliproxy/pkg/queue"
	"reliproxy/pkg/repository"
	"reliproxy/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

func main() {
	// サーキットブレーカーの設定
	circuitBreaker := initCircuitBreaker()
	rateLimiter := rate.NewLimiter(rate.Limit(5), 10)
	httpClient := &httpclient.DefaultHttpClient{}

	reliClient := httpclient.NewReliClient(httpClient, circuitBreaker, rateLimiter, 3)

	handler := handlers.NewHandler(reliClient)

	dbn, err := initDatabase()
	if err != nil {
		panic(err)
	}
	statusRepository := repository.NewGormRequestStatusRepository(dbn)

	rdb := initRedisClient()
	queue := queue.NewRedisQueue(rdb, utils.GetEnv("REDIS_QUEUE_NAME", "queue"))

	asyncWriteHandler := handlers.NewAsyncWriteHandler(queue, statusRepository)

	consumer := consumer.NewConsumer(queue, statusRepository, reliClient)
	go consumer.Start()

	// Ginルーターの設定
	r := gin.Default()
	r.GET("/proxy", handler.HandleRequest)
	r.GET("/async-proxy", asyncWriteHandler.HandleRequest)

	// サーバーの起動
	r.Run(":8080")
}

func initCircuitBreaker() *gobreaker.CircuitBreaker {
	cbSettings := gobreaker.Settings{
		Name:        "HTTP GET",
		MaxRequests: 5,
		Interval:    2 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.TotalFailures > 3
		},
	}
	return gobreaker.NewCircuitBreaker(cbSettings)
}

func initDatabase() (*gorm.DB, error) {
	connectionEnv := db.NewMySQLConnectionEnv()
	dbn, err := connectionEnv.ConnectDBWithRetry()
	if err != nil {
		panic(err)
	}
	if err := db.Migrate(dbn); err != nil {
		return nil, err
	}
	return dbn, nil
}

func initRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     utils.GetEnv("REDIS_ADDR", "localhost:6379"),
		Password: utils.GetEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})
}
