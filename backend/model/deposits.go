package model

import (
	"web3-wallet-exchange/global"

	"gorm.io/gorm"
)

const (
	DepositPending   = "pending"
	DepositConfirmed = "confirmed"
)

// Deposit 充值流水：一笔链上转入对应一行，状态机载体（pending → confirmed）。
// 幂等键是 TxHash（ETH 转账用 tx_hash 即可；ERC20 阶段再考虑 log_index）。
type Deposit struct {
	gorm.Model
	UserID        uint   `gorm:"index;not null;comment:入账用户ID"`
	TxHash        string `gorm:"uniqueIndex;size:66;not null;comment:链上交易哈希，幂等键"`
	FromAddress   string `gorm:"size:42;comment:转出地址"`
	ToAddress     string `gorm:"size:42;index;not null;comment:收款地址，应等于用户托管充值地址"`
	Amount        string `gorm:"type:decimal(36,0);not null;comment:金额，单位 wei 整数字符串"`
	Asset         string `gorm:"size:16;not null;default:ETH;comment:资产符号，阶段3为 ETH"`
	Status        string `gorm:"size:32;not null;index;comment:pending=已见交易 confirmed=已入账"`
	BlockNumber   uint64 `gorm:"not null;comment:交易所在区块号"`
	Confirmations uint   `gorm:"not null;default:0;comment:当前确认数，扫链时更新"`
}

func AutoMigrateDeposit() error {
	return global.DB.AutoMigrate(&Deposit{})
}
