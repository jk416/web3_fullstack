package handler

import (
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetWallet(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userID.(uint)
	wallet, err := service.GetOrCreateWallet(userId)
	if err != nil {
		global.Log.Error("get wallet failed", zap.Error(err))
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(200, gin.H{"address": wallet.Address})
}
