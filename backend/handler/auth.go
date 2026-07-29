package handler

import (
	"web3-wallet-exchange/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
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
		c.JSON(400, gin.H{"error": err.Error()})
	} else {
		c.JSON(200, gin.H{"nonce": nonce})
	}
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
	if err := service.VerifyLogin(req.Address, req.Signature); err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Login successful"})
}
