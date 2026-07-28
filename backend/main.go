package main

import (
	"fmt"
	"log"
	"web3-wallet-exchange/global"
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
	initRouter := router.InitRouter()
	global.Log.Info("server started", zap.Int("port", global.Conf.Server.Port))
	if err := initRouter.Run(fmt.Sprintf(":%d", global.Conf.Server.Port)); err != nil {
		global.Log.Fatal("server failed to start", zap.Error(err))
	}
}
