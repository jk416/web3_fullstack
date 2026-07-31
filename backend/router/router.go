package router

import (
	"web3-wallet-exchange/handler"
	"web3-wallet-exchange/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	auth := r.Group("/api/auth")
	auth.GET("/nonce", handler.GetNonce)
	auth.POST("/login", handler.Login)
	authed := r.Group("/api", middleware.JWTAuth())
	authed.GET("/me", handler.Me)
	r.GET("/api/health", handler.Health)
	return r
}
