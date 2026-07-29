package router

import (
	"web3-wallet-exchange/handler"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	auth := r.Group("/api/auth")
	auth.GET("/nonce", handler.GetNonce)
	auth.POST("/api/login", handler.Login)
	r.GET("/api/health", handler.Health)
	return r
}
