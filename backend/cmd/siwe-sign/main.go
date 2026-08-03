// 本地测试工具：模拟一个钱包走完 SIWE 登录（取 nonce → 签名 → 换 token）。
// 仅用于开发期手测；绝不要填入真实主网私钥。
//
// 运行（后端在 :8080）：
//
//	go run ./cmd/siwe-sign
//
// 默认每次随机一把登录钥 → 新 users 行 → 新托管充值地址。
// 要复用同一个登录身份（从而复用同一 user_id / 充值地址），固定私钥：
//
//	export SIWE_TEST_KEY=<64位hex，可带0x>
//	go run ./cmd/siwe-sign
//
// 首次随机生成时会打印可 export 的 SIWE_TEST_KEY，复制后再跑即可稳定复用。
package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

const base = "http://localhost:8080"

func main() {
	priv, reused, err := loadOrCreateTestKey()
	if err != nil {
		log.Fatal(err)
	}
	addr := crypto.PubkeyToAddress(priv.PublicKey).Hex()

	// 1) 取 nonce
	nonce, err := getNonce(addr)
	if err != nil {
		log.Fatal(err)
	}

	// 2) 按后端契约构造消息，并按 EIP-191 方式哈希后签名
	msg := "Sign in to Web3 Wallet Exchange.\n\nNonce: " + nonce
	hash := accounts.TextHash([]byte(msg))
	sig, err := crypto.Sign(hash, priv) // v = 0/1
	if err != nil {
		log.Fatal(err)
	}
	sig[64] += 27 // 模拟 MetaMask 返回的 v = 27/28
	sigHex := "0x" + hex.EncodeToString(sig)

	// 3) 登录换 token
	token, err := login(addr, sigHex)
	if err != nil {
		log.Fatal(err)
	}

	// 4) 顺手打一下托管充值地址，方便确认「同一登录钥 → 同一 deposit address」
	deposit, err := getDepositAddress(token)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("login_address  :", addr)
	if reused {
		fmt.Println("key_source     : SIWE_TEST_KEY (reused)")
	} else {
		fmt.Println("key_source     : random (new user each run unless you export the key below)")
		fmt.Printf("reuse_next_time: export SIWE_TEST_KEY=%s\n", hex.EncodeToString(crypto.FromECDSA(priv)))
	}
	fmt.Println("token          :", token)
	fmt.Println("deposit_address:", deposit)
}

// loadOrCreateTestKey：有 SIWE_TEST_KEY 则复用；否则生成一把仅用于本地的随机钥。
func loadOrCreateTestKey() (priv *ecdsa.PrivateKey, reused bool, err error) {
	raw := strings.TrimSpace(os.Getenv("SIWE_TEST_KEY"))
	if raw == "" {
		priv, err = crypto.GenerateKey()
		return priv, false, err
	}
	raw = strings.TrimPrefix(raw, "0x")
	priv, err = crypto.HexToECDSA(raw)
	if err != nil {
		return nil, false, fmt.Errorf("SIWE_TEST_KEY invalid: %w", err)
	}
	return priv, true, nil
}

func getNonce(addr string) (string, error) {
	resp, err := http.Get(base + "/api/auth/nonce?address=" + url.QueryEscape(addr))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out struct {
		Nonce string `json:"nonce"`
	}
	if err := json.Unmarshal(body, &out); err != nil || out.Nonce == "" {
		return "", fmt.Errorf("get nonce failed: %s", body)
	}
	return out.Nonce, nil
}

func login(addr, sig string) (string, error) {
	payload, _ := json.Marshal(map[string]string{"address": addr, "signature": sig})
	resp, err := http.Post(base+"/api/auth/login", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &out); err != nil || out.Token == "" {
		return "", fmt.Errorf("login failed: %s", body)
	}
	return out.Token, nil
}

// getDepositAddress 调受保护接口 GET /api/wallet（与 handler 返回字段 walletAddress 对齐）。
func getDepositAddress(token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, base+"/api/wallet", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out struct {
		// handler 当前字段名；若以后改成 address 可两处兼容
		WalletAddress string `json:"walletAddress"`
		Address       string `json:"address"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("get wallet failed: %s", body)
	}
	if out.WalletAddress != "" {
		return out.WalletAddress, nil
	}
	if out.Address != "" {
		return out.Address, nil
	}
	return "", fmt.Errorf("get wallet failed: %s", body)
}
