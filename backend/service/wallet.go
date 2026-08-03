package service

import (
	"errors"
	"strings"
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/model"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func GetOrCreateWallet(userID uint) (*model.Wallet, error) {
	keyLen := len(global.Conf.Wallet.EncryptionKey)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, errors.New("invalid wallet encryption key size")
	}
	var wallet model.Wallet
	result := global.DB.Where("user_id = ?", userID).First(&wallet)

	if result.Error == nil {
		return &wallet, nil
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error // 真系统错误，别假装去建
	}
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	addr := crypto.PubkeyToAddress(key.PublicKey) // common.Address
	address := strings.ToLower(addr.Hex())
	plain := crypto.FromECDSA(key) // []byte, 32
	cipherText, err := Encrypt(plain, []byte(global.Conf.Wallet.EncryptionKey))
	if err != nil {
		return nil, err
	}
	wallet = model.Wallet{
		UserID:              userID,
		Address:             address,
		EncryptedPrivateKey: cipherText,
	}
	result = global.DB.Create(&wallet)
	if result.Error != nil {
		// 判断错误类型是否为 Postgres 的 PgError，且错误码为 23505 (unique_violation)
		if pgErr, ok := errors.AsType[*pgconn.PgError](result.Error); ok && pgErr.Code == "23505" {
			var existing model.Wallet
			if err := global.DB.Where("user_id = ?", userID).First(&existing).Error; err != nil {
				return nil, err
			}
			return &existing, nil
		}
		return nil, result.Error
	}
	return &wallet, nil

}
