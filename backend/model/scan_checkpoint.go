package model

import (
	"web3-wallet-exchange/global"

	"gorm.io/gorm"
)

// ScanCheckpoint 扫链水位。教学实现只维护 id=1 一行。
// LastBlock = 已成功处理完的最大块号；下一轮从 LastBlock+1 扫到 head。
type ScanCheckpoint struct {
	gorm.Model
	LastBlock uint64 `gorm:"not null;comment:已成功处理完的区块号"`
}

func AutoMigrateScanCheckpoint() error {
	return global.DB.AutoMigrate(&ScanCheckpoint{})
}
