package model

import (
	"web3-wallet-exchange/global"

	"gorm.io/gorm"
)

type Balance struct {
	gorm.Model
	UserID uint   `gorm:"uniqueIndex:uk_user_asset;not null;comment:用户ID"`
	Asset  string `gorm:"uniqueIndex:uk_user_asset;size:16;not null;comment:资产符号，如 ETH"`
	// 金额：DB 用 DECIMAL，Go 用 string 存整数字符串（wei）
	Available string `gorm:"type:decimal(36,0);not null;default:0;comment:可用余额，单位 wei"`
}

func AutoMigrateBalance() error {
	return global.DB.AutoMigrate(&Balance{})
}
