package main

import (
	"telegram-premium/pkg/bootstrap"
	"telegram-premium/pkg/core/cst"
)

func main() {
	bootstrap.LoadConfig(cst.AppName)
	bootstrap.ConnectDB()
	err := bootstrap.Telegram()
	if err != nil {
		panic(err)
	}
}
