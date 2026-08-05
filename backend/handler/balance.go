package handler

import (
	"net/http"
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/model"
	"web3-wallet-exchange/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetBalances(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userID.(uint)
	balances, err := service.ListBalances(userId)
	if err != nil {
		global.Log.Error("list balances failed", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	// 保证 JSON 是 [] 而不是 null
	if balances == nil {
		balances = []model.Balance{}
	}
	c.JSON(http.StatusOK, gin.H{"balances": balances})
}
