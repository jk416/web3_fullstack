package main

import (
	"fmt"
	"log"
	"web3-wallet-exchange/global"
	"web3-wallet-exchange/router"
)

func main() {
	err := global.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	initRouter := router.InitRouter()
	err = initRouter.Run(fmt.Sprintf(":%d", global.Conf.Server.Port))
	if err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
