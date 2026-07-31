package handler

import (
	"errors"
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LoginRequest struct {
	Address   string `json:"address"  binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

func GetNonce(c *gin.Context) {
	addr := c.Query("address")
	if addr == "" {
		c.JSON(400, gin.H{"error": "Address is empty"})
		return
	}
	if !common.IsHexAddress(addr) {
		c.JSON(400, gin.H{"error": "Invalid address"})
		return
	}
	nonce, err := service.GenerateNonce(addr)
	if err != nil {
		global.Log.Error("generate nonce failed", zap.Error(err))
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(200, gin.H{"nonce": nonce})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if !common.IsHexAddress(req.Address) {
		c.JSON(400, gin.H{"error": "Invalid address"})
		return
	}
	userID, err := service.VerifyLogin(req.Address, req.Signature)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidSignature):
			c.JSON(401, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(400, gin.H{"error": err.Error()})
		default: // 未知 = 系统错：记日志，不把细节泄给客户端
			global.Log.Error("login failed", zap.Error(err))
			c.JSON(500, gin.H{"error": "internal server error"})
		}
		return
	}
	tokenStr, err := global.Token.CreateToken(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create token"})
		return
	}
	c.JSON(200, gin.H{"token": tokenStr})
}

func Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "userID not found"})
		return
	}
	c.JSON(200, gin.H{"userID": userID})
}
