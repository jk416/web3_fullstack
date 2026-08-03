package model

import (
	"web3-wallet-exchange/global"

	"gorm.io/gorm"
)

type Wallet struct {
	gorm.Model
	UserID              uint   `gorm:"uniqueIndex;not null"`
	Address             string `gorm:"uniqueIndex;size:42;not null"`
	EncryptedPrivateKey string `gorm:"not null"`
}

func AutoMigrateWallet() error {
	return global.DB.AutoMigrate(&Wallet{})
}
