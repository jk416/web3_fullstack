package main

import (
	"fmt"
	"log"
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/model"
	"web3-wallet-exchange/router"

	"go.uber.org/zap"
)

func main() {
	err := global.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	err = global.InitLogger()
	defer global.Log.Sync()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	err = global.InitDB()
	if err != nil {
		global.Log.Fatal("Failed to initialize DB", zap.Error(err))
	}
	err = model.AutoMigrateUser()
	if err != nil {
		global.Log.Fatal("auto migrate failed", zap.Error(err))
	}
	global.Log.Info("database connected")
	initRouter := router.InitRouter()
	global.Log.Info("server starting", zap.Int("port", global.Conf.Server.Port))
	if err := initRouter.Run(fmt.Sprintf(":%d", global.Conf.Server.Port)); err != nil {
		global.Log.Fatal("server failed to start", zap.Error(err))
	}
}
