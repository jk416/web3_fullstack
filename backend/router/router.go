package router

import (
	"web3-wallet-exchange/handler"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/api/health", handler.Health)
	return r
}
