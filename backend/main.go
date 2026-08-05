package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		service.RunScanner(ctx)
	}()
	go func() {
		defer wg.Done()
		service.RunConfirmScanner(ctx)
	}()

	global.Log.Info("database connected")
	r := router.InitRouter()
	global.Log.Info("server starting", zap.Int("port", global.Conf.Server.Port))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", global.Conf.Server.Port),
		Handler: r, // InitRouter() 返回的 engine
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			global.Log.Fatal("server failed", zap.Error(err))
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	global.Log.Info("Shutdown signal received, exiting...")
	cancel()

	// 给 HTTP 一个关机超时（如 10s）
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx) // 等进行中的请求结束（有限时）

	global.Log.Info("server shutdown")

	// 信号后 cancel + Shutdown HTTP 之后：
	wg.Wait() // 或带超时的 wait（select + time.After），教学用 Wait 即可
	global.Log.Info("bye")

}
