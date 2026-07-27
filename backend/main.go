package main

import (
	"log"
	"web3-wallet-exchange/router"
)

func main() {
	initRouter := router.InitRouter()
	err := initRouter.Run("localhost:8080")
	if err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
