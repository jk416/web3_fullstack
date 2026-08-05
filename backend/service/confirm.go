package service

import (
	"context"
	"time"
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ctx 取消后退出（main 里 defer cancel；阶段 3.5 再接信号优雅退出）。
func RunConfirmScanner(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			global.Log.Error("panic", zap.Any("error", r))
		}
	}()
	sec := global.Conf.Ethereum.ScanIntervalSec
	if sec <= 0 {
		sec = 5
	}
	ticker := time.NewTicker(time.Duration(sec) * time.Second)
	defer ticker.Stop()

	global.Log.Info("confirm scanner started",
		zap.Int("interval_sec", sec),
		zap.String("rpc", global.Conf.Ethereum.RPCURL),
	)

	// 启动先跑一轮，不用干等第一个 tick
	if err := confirmOnce(ctx); err != nil {
		global.Log.Error("confirm once failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			global.Log.Info("confirm scanner stopped")
			return
		case <-ticker.C:
			if err := confirmOnce(ctx); err != nil {
				global.Log.Error("confirm once failed", zap.Error(err))
			}
		}
	}
}

func confirmOnce(ctx context.Context) error {
	head, err := global.EthClient.BlockNumber(ctx)
	if err != nil {
		return err
	}
	var deposits []model.Deposit
	err = global.DB.Where("status = ?", model.DepositPending).Find(&deposits).Error
	if err != nil {
		return err
	}
	for _, deposit := range deposits {
		if head < deposit.BlockNumber {
			continue
		}
		var conf = int(head - deposit.BlockNumber + 1)
		need := global.Conf.Ethereum.Confirmations
		if need <= 0 {
			need = 1
		}
		if conf < need {
			tx := global.DB.Where("id = ?", deposit.ID).Update("confirmations", conf)
			if tx.Error != nil {
				continue
			}
			continue
		}
		err := global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			res := tx.Model(&model.Deposit{}).
				Where("id = ? AND status = ?", deposit.ID, model.DepositPending).
				Updates(map[string]interface{}{
					"status":        model.DepositConfirmed,
					"confirmations": conf,
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				// 没有更新到记录，说明状态发生变更
				return nil
			}

			// 1) 确保有余额行；没有则插入 Available=0（Attrs 只在 INSERT 时带上）
			bal := model.Balance{
				UserID: deposit.UserID,
				Asset:  deposit.Asset,
			}
			if err := tx.Where(model.Balance{UserID: deposit.UserID, Asset: deposit.Asset}).
				Attrs(model.Balance{Available: "0"}).
				FirstOrCreate(&bal).Error; err != nil {
				return err
			}

			walletRes := tx.Model(&model.Balance{}).
				Where("user_id = ? AND asset = ?", deposit.UserID, deposit.Asset).
				Update("available", gorm.Expr("available + ?", deposit.Amount))

			if walletRes.Error != nil {
				return walletRes.Error
			}
			global.Log.Info("confirm deposit success",
				zap.Uint("deposit_id", deposit.ID),
				zap.String("tx_hash", deposit.TxHash))
			return nil
		})
		if err != nil {
			global.Log.Error("confirm deposit failed",
				zap.Uint("deposit_id", deposit.ID),
				zap.String("tx_hash", deposit.TxHash),
				zap.Error(err))
			continue
		}
	}
	return nil
}
