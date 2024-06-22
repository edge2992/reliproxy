package main

import (
	"third-party-proxy/pkg/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Ginルーターの設定
	r := gin.Default()
	r.GET("/proxy", handlers.HandleRequest)
	r.Run(":8080")
}
