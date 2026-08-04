package global

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

// EthClient 全局以太坊 JSON-RPC 客户端（Sepolia 等），扫链用。
var EthClient *ethclient.Client

func InitEthClient() error {
	if Conf.Ethereum.RPCURL == "" {
		return fmt.Errorf("ethereum.rpc_url is empty")
	}
	cli, err := ethclient.Dial(Conf.Ethereum.RPCURL)
	if err != nil {
		return fmt.Errorf("dial eth rpc: %w", err)
	}
	EthClient = cli
	return nil
}
