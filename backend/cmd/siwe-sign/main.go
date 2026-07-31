// 本地测试工具：模拟一个钱包走完 SIWE 登录（取 nonce → 签名 → 换 token）。
// 仅用于开发期手测，等 1.6 前端接好后就由浏览器钱包完成这套。
// 运行：后端起在 :8080 时，go run ./cmd/siwe-sign
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

const base = "http://localhost:8080"

func main() {
	// 每次生成一个全新测试私钥（绝不能用真钱包私钥）
	priv, err := crypto.GenerateKey()
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

	fmt.Println("address  :", addr)
	fmt.Println("nonce    :", nonce)
	fmt.Println("signature:", sigHex)
	fmt.Println("token    :", token)
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
