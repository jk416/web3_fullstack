package main

import (
	"context"
	"fmt"
	"log"
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/model"
	"web3-wallet-exchange/router"
	"web3-wallet-exchange/service"

	"go.uber.org/zap"
)

func main() {
	if err := global.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if err := global.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer global.Log.Sync()

	if err := global.InitDB(); err != nil {
		global.Log.Fatal("Failed to initialize DB", zap.Error(err))
	}
	global.InitToken()

	// 表结构：教学阶段仍用 AutoMigrate；生产应改为 versioned migration
	if err := model.AutoMigrateUser(); err != nil {
		global.Log.Fatal("auto migrate user failed", zap.Error(err))
	}
	if err := model.AutoMigrateWallet(); err != nil {
		global.Log.Fatal("auto migrate wallet failed", zap.Error(err))
	}
	if err := model.AutoMigrateBalance(); err != nil {
		global.Log.Fatal("auto migrate balance failed", zap.Error(err))
	}
	if err := model.AutoMigrateDeposit(); err != nil {
		global.Log.Fatal("auto migrate deposit failed", zap.Error(err))
	}
	if err := model.AutoMigrateScanCheckpoint(); err != nil {
		global.Log.Fatal("auto migrate scan_checkpoint failed", zap.Error(err))
	}

	// 扫链依赖 RPC，必须在 goroutine 之前初始化
	if err := global.InitEthClient(); err != nil {
		global.Log.Fatal("failed to initialize eth client", zap.Error(err))
	}

	// 后台扫链（阶段 3.5 再接 SIGINT 优雅退出）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.RunScanner(ctx)

	global.Log.Info("database connected")
	r := router.InitRouter()
	global.Log.Info("server starting", zap.Int("port", global.Conf.Server.Port))
	if err := r.Run(fmt.Sprintf(":%d", global.Conf.Server.Port)); err != nil {
		global.Log.Fatal("server failed to start", zap.Error(err))
	}
}
