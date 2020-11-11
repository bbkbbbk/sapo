package main

import (
	"os"

	"github.com/labstack/echo"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"

	pkgMongo "github.com/bbkbbbk/sapo/pkg/mongo"
	"github.com/bbkbbbk/sapo/server"
)

var (
	bot     *linebot.Client
	db      *mongo.Database
	spotify server.SpotifyService
	err     error
)

func init() {
	bot, err = linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Err(err)
	}
}

func init() {
	db = pkgMongo.NewMongo(pkgMongo.Config{
		AuthSource: os.Getenv("MONGO_AUTH_SOURCE"),
		Database:   os.Getenv("MONGO_DATABASE"),
		Host:       os.Getenv("MONGO_HOST"),
		Username:   os.Getenv("MONGO_USERNAME"),
		Password:   os.Getenv("MONGO_PASSWORD"),
	})
}

func main() {
	e := echo.New()

	serverHandler := server.NewHandler(bot, spotify)
	repository := server.NewRepository(db)
	spotify = server.NewSpotifyService(
		os.Getenv("MY_CLIENT_ID"),
		os.Getenv("MY_CLIENT_SECRET"),
		repository,
	)

	e.GET("/", serverHandler.HomePage)
	e.GET("/ping", serverHandler.PingCheck)
	e.POST("/callback", serverHandler.Callback)

	port := ":" + os.Getenv("APP_PORT")
	e.Logger.Fatal(e.Start(port))
}
