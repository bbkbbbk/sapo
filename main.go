package main

import (
	"log"
	"os"

	"github.com/labstack/echo"
	"github.com/line/line-bot-sdk-go/linebot"

	"github.com/bbkbbbk/sapo/server"
)

var (
	bot *linebot.Client
	err error
)

func init() {
	bot, err = linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	e := echo.New()

	serverHandler := server.NewHandler(bot)

	e.GET("/ping", serverHandler.PingCheck)
	e.POST("/callback", serverHandler.Callback)

	port := ":" + os.Getenv("APP_PORT")
	e.Logger.Fatal(e.Start(port))
}
