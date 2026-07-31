package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/model"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound     = errors.New("user not found, request a nonce first")
	ErrInvalidSignature = errors.New("invalid signature")
)

func GenerateNonce(walletAddress string) (string, error) {
	addr := strings.ToLower(walletAddress)
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(b)
	//查询是否有该地址用户，没有就创建，有就更新
	var user model.User
	result := global.DB.Where(model.User{WalletAddress: addr}).Assign(model.User{Nonce: nonce}).FirstOrCreate(&user)
	if result.Error != nil {
		return "", result.Error
	}
	return nonce, nil
}

func VerifyLogin(walletAddress, signature string) (uint, error) {
	addr := strings.ToLower(walletAddress)
	var user model.User
	result := global.DB.Where(&model.User{WalletAddress: addr}).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, ErrUserNotFound
		}
		return 0, result.Error // 其他 DB 错误 = 系统错，原样上抛
	}
	msg := "Sign in to Web3 Wallet Exchange.\n\nNonce: " + user.Nonce
	sig := common.FromHex(signature)
	if sig == nil {
		return 0, ErrInvalidSignature
	}
	if len(sig) != 65 {
		return 0, ErrInvalidSignature
	}
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	hash := accounts.TextHash([]byte(msg))
	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return 0, ErrInvalidSignature
	}
	recovered := crypto.PubkeyToAddress(*pubKey)
	if strings.ToLower(recovered.Hex()) != addr {
		return 0, ErrInvalidSignature
	}
	_, err = GenerateNonce(addr)
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}
