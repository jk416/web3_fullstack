package handler

import (
	"errors"
	"net/http"
	"strconv"
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetDeposits
//
//	GET /api/deposits?page=1&page_size=10&status=pending&asset=ETH&tx_hash=0xabc
//
// 传统列表：分页 + 多条件查询；只返回当前登录用户的数据。
func GetDeposits(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userID.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := service.ListDeposits(userId, service.DepositListQuery{
		Page:     page,
		PageSize: pageSize,
		Status:   c.Query("status"),
		Asset:    c.Query("asset"),
		TxHash:   c.Query("tx_hash"),
	})
	if err != nil {
		global.Log.Error("list deposits failed", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 字段命名稳定，方便前端/对接文档（类 PageResult）
	c.JSON(http.StatusOK, gin.H{
		"list":      result.Items,
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.Size,
	})
}

// GetDepositDetail GET /api/deposits/:id
func GetDepositDetail(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userID.(uint)

	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	d, err := service.GetDepositByID(userId, uint(id64))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "deposit not found"})
			return
		}
		global.Log.Error("get deposit failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deposit": d})
}
