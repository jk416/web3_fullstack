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

	// 需登录：JWT 中间件写入 userID
	authed := r.Group("/api", middleware.JWTAuth())
	authed.GET("/me", handler.Me)
	authed.GET("/wallet", handler.GetWallet)
	authed.GET("/balances", handler.GetBalances)
	// 列表（分页+条件）与详情：注意 :id 写在具体路径旁，Gin 按注册匹配
	authed.GET("/deposits", handler.GetDeposits)
	authed.GET("/deposits/:id", handler.GetDepositDetail)

	r.GET("/api/health", handler.Health)
	return r
}
