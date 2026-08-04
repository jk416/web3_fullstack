package service

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/model"

	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 每轮最多处理的块数，避免公共 RPC 一次打爆、也避免单 tick 过久。
const maxBlocksPerTick = 20

// RunScanner 后台扫链循环：按配置间隔调用 scanOnce。
// ctx 取消后退出（main 里 defer cancel；阶段 3.5 再接信号优雅退出）。
func RunScanner(ctx context.Context) {
	sec := global.Conf.Ethereum.ScanIntervalSec
	if sec <= 0 {
		sec = 5
	}
	ticker := time.NewTicker(time.Duration(sec) * time.Second)
	defer ticker.Stop()

	global.Log.Info("scanner started",
		zap.Int("interval_sec", sec),
		zap.String("rpc", global.Conf.Ethereum.RPCURL),
	)

	// 启动先跑一轮，不用干等第一个 tick
	if err := scanOnce(ctx); err != nil {
		global.Log.Error("scan once failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			global.Log.Info("scanner stopped")
			return
		case <-ticker.C:
			if err := scanOnce(ctx); err != nil {
				global.Log.Error("scan once failed", zap.Error(err))
			}
		}
	}
}

// scanOnce 一轮：读水位 → 从 LastBlock+1 扫到 head（限块数）→ 每块成功后推进水位。
func scanOnce(ctx context.Context) error {
	if global.EthClient == nil {
		return errors.New("eth client not initialized")
	}

	addrMap, err := loadDepositAddressMap()
	if err != nil {
		return err
	}
	// 还没有任何托管地址：空转即可，不算错误
	if len(addrMap) == 0 {
		return nil
	}

	head, err := global.EthClient.BlockNumber(ctx)
	if err != nil {
		return err
	}

	from, err := getOrInitCheckpoint(head)
	if err != nil {
		return err
	}
	if from > head {
		// 已追上最新块
		return nil
	}

	to := head
	if to > from+maxBlocksPerTick-1 {
		to = from + maxBlocksPerTick - 1
	}

	chainID, err := global.EthClient.NetworkID(ctx)
	if err != nil {
		return err
	}

	global.Log.Debug("scanning blocks",
		zap.Uint64("from", from),
		zap.Uint64("to", to),
		zap.Uint64("head", head),
		zap.Int("wallets", len(addrMap)),
	)

	for n := from; n <= to; n++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := processBlock(ctx, n, chainID, addrMap); err != nil {
			// 本块失败：不推进水位，下轮从同一 from 重试
			return err
		}
		if err := advanceCheckpoint(n); err != nil {
			return err
		}
	}
	return nil
}

// loadDepositAddressMap 托管充值地址 → userID（key 一律小写）。
func loadDepositAddressMap() (map[string]uint, error) {
	var wallets []model.Wallet
	// Find 无数据时 err==nil 且 slice 为空，不是 ErrRecordNotFound
	if err := global.DB.Find(&wallets).Error; err != nil {
		return nil, err
	}
	m := make(map[string]uint, len(wallets))
	for _, w := range wallets {
		m[strings.ToLower(w.Address)] = w.UserID
	}
	return m, nil
}

// resolveStartBlock 第一次无水位时的起跑线。
// 配置 start_block>0 用配置；否则 head-20（head 不足 20 则从 0）。
func resolveStartBlock(head uint64) uint64 {
	if global.Conf.Ethereum.StartBlock > 0 {
		return global.Conf.Ethereum.StartBlock
	}
	if head >= 20 {
		return head - 20
	}
	return 0
}

// getOrInitCheckpoint 返回本轮应开始扫描的块号 from。
//
//	已有 id=1：from = LastBlock + 1
//	没有：按 start 建档，LastBlock = start-1（start==0 时 LastBlock=0 且 from=0），from = start
func getOrInitCheckpoint(head uint64) (uint64, error) {
	var cp model.ScanCheckpoint
	err := global.DB.First(&cp, 1).Error
	if err == nil {
		return cp.LastBlock + 1, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	start := resolveStartBlock(head)
	var last uint64
	if start > 0 {
		last = start - 1
	} else {
		last = 0
	}

	cp = model.ScanCheckpoint{
		LastBlock: last,
	}
	cp.ID = 1
	if err := global.DB.Create(&cp).Error; err != nil {
		// 并发下可能已被别人创建：再读一次
		var again model.ScanCheckpoint
		if e2 := global.DB.First(&again, 1).Error; e2 != nil {
			return 0, err
		}
		return again.LastBlock + 1, nil
	}

	// start==0 时 last=0，若 return last+1 会跳过块 0；应返回 start
	return start, nil
}

func advanceCheckpoint(blockNum uint64) error {
	return global.DB.Model(&model.ScanCheckpoint{}).
		Where("id = ?", 1).
		Update("last_block", blockNum).Error
}

// processBlock 处理单个块：To 命中托管地址且 Value>0 的 ETH 转账 → 幂等写入 deposits(pending)。
func processBlock(ctx context.Context, blockNum uint64, chainID *big.Int, addrMap map[string]uint) error {
	block, err := global.EthClient.BlockByNumber(ctx, new(big.Int).SetUint64(blockNum))
	if err != nil {
		return err
	}

	signer := types.LatestSignerForChainID(chainID)

	for _, tx := range block.Transactions() {
		to := tx.To()
		if to == nil {
			continue // 合约创建
		}
		toAddr := strings.ToLower(to.Hex())
		userID, ok := addrMap[toAddr]
		if !ok {
			continue
		}
		if tx.Value().Sign() == 0 {
			continue // 无 ETH 转账（可能是合约调用）
		}

		fromAddr, err := types.Sender(signer, tx)
		if err != nil {
			global.Log.Error("recover tx sender failed",
				zap.Uint64("block", blockNum),
				zap.String("tx", tx.Hash().Hex()),
				zap.Error(err),
			)
			continue
		}

		dep := model.Deposit{
			UserID:        userID,
			TxHash:        strings.ToLower(tx.Hash().Hex()),
			FromAddress:   strings.ToLower(fromAddr.Hex()),
			ToAddress:     toAddr,
			Amount:        tx.Value().String(),
			Asset:         "ETH",
			Status:        model.DepositPending,
			BlockNumber:   blockNum,
			Confirmations: 0,
		}
		// tx_hash 唯一：重复扫同一笔 DoNothing
		if err := global.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&dep).Error; err != nil {
			return err
		}
		global.Log.Info("deposit pending observed",
			zap.Uint("user_id", userID),
			zap.String("tx", dep.TxHash),
			zap.String("to", toAddr),
			zap.String("amount_wei", dep.Amount),
			zap.Uint64("block", blockNum),
		)
	}
	return nil
}
