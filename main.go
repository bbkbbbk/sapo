package main

import (
	pkgMongo "github.com/bbkbbbk/sapo/pkg/mongo"
	"os"

	"github.com/labstack/echo"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bbkbbbk/sapo/server"
)

var (
	db      *mongo.Database
	line    *linebot.Client
	spotify server.SpotifyService
	err     error
)

func init() {
	db = pkgMongo.NewMongo(pkgMongo.Config{
		AuthSource: os.Getenv("MONGO_AUTH_SOURCE"),
		Database:   os.Getenv("MONGO_DATABASE"),
		Host:       os.Getenv("MONGO_HOST"),
		Username:   os.Getenv("MONGO_USERNAME"),
		Password:   os.Getenv("MONGO_PASSWORD"),
	})
}

func init() {
	line, err = linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		logrus.Warnf("[init]: unable to initialize line line client %v", err)
	}
}

func init() {
	spotify = server.NewSpotifyService(
		os.Getenv("MY_CLIENT_ID"),
		os.Getenv("MY_CLIENT_SECRET"),
	)
}

func main() {
	e := echo.New()

	repository := server.NewRepository(db)
	service := server.NewService(line, spotify, repository)
	serverHandler := server.NewHandler(service)
	server.RoutesRegister(e, serverHandler)

	port := ":" + os.Getenv("APP_PORT")
	e.Logger.Fatal(e.Start(port))
}
