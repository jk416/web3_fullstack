package model

import (
	"web3-wallet-exchange/global"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	WalletAddress string `gorm:"uniqueIndex;size:42;not null" json:"wallet_address"`
	Nonce         string `gorm:"not null" json:"nonce"`
}

func AutoMigrateUser() error {
	return global.DB.AutoMigrate(&User{})
}
