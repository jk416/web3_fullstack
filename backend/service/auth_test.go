package service

import (
	"encoding/hex"
	"os"
	"testing"

	"web3-wallet-exchange/global"
	"web3-wallet-exchange/model"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

// 端到端自测：用生成的私钥模拟 MetaMask 签名，验证 VerifyLogin 的恢复逻辑 + 防重放。
// 需要 postgres 在跑（localhost:5433）。
func TestVerifyLogin(t *testing.T) {
	// go test 的工作目录是本包目录(backend/service)，切回 backend 让 config/ 相对路径可用。
	if err := os.Chdir(".."); err != nil {
		t.Fatal(err)
	}
	if err := global.LoadConfig(); err != nil {
		t.Fatalf("load config: %v", err)
	}
	if err := global.InitLogger(); err != nil {
		t.Fatalf("init logger: %v", err)
	}
	if err := global.InitDB(); err != nil {
		t.Skipf("no DB available, skipping: %v", err)
	}
	if err := model.AutoMigrateUser(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// 造一个私钥 + 地址
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	addr := crypto.PubkeyToAddress(key.PublicKey).Hex()

	// 走一遍 nonce 发放（创建/更新用户，拿到当前 nonce）
	nonce, err := GenerateNonce(addr)
	if err != nil {
		t.Fatal(err)
	}

	// 用和后端完全一致的消息契约，按 EIP-191 方式哈希后签名
	msg := "Sign in to Web3 Wallet Exchange.\n\nNonce: " + nonce
	hash := accounts.TextHash([]byte(msg))
	sig, err := crypto.Sign(hash, key) // v 为 0/1
	if err != nil {
		t.Fatal(err)
	}
	sig[64] += 27 // 模拟 MetaMask 返回 27/28
	sigHex := "0x" + hex.EncodeToString(sig)

	// 1) 正确签名应通过，并拿到 userID
	userID, err := VerifyLogin(addr, sigHex)
	if err != nil {
		t.Fatalf("expected valid login, got: %v", err)
	}

	// 1b) JWT 闭环：签发的 token 应能被解回同一个 userID
	global.InitToken()
	tok, err := global.Token.CreateToken(userID)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	claims, err := global.Token.VerifyToken(tok)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("token userID mismatch: got %d, want %d", claims.UserID, userID)
	}

	// 2) 重放同一签名应失败（登录成功后 nonce 已轮换，消息变了）
	if _, err := VerifyLogin(addr, sigHex); err == nil {
		t.Fatal("expected replay to FAIL after nonce rotation, but it passed")
	}
}
