package service

import (
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/model"
)

func ListBalances(userId uint) ([]model.Balance, error) {
	var balance []model.Balance
	tx := global.DB.Where("user_id = ?", userId).Find(&balance)
	if tx.Error != nil {
		global.Log.Error(tx.Error.Error())
		return nil, tx.Error
	}
	return balance, nil
}
