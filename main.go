package main

import (
	"github.com/labstack/echo/middleware"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/mongo"

	pkgMongo "github.com/bbkbbbk/sapo/pkg/mongo"
	"github.com/bbkbbbk/sapo/server"
)

var (
	db      *mongo.Database
	line    server.LINEService
	spotify server.SpotifyService
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
	line = server.NewLINEService(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
}

func init() {
	spotify = server.NewSpotifyService(
		os.Getenv("MY_CLIENT_ID"),
		os.Getenv("MY_CLIENT_SECRET"),
	)
}

func main() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{os.Getenv("CORS_ALLOW_ORIGIN")},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodPut},
	}))

	repository := server.NewRepository(db)
	service := server.NewService(line, spotify, repository)
	serverHandler := server.NewHandler(service)
	server.RoutesRegister(e, serverHandler)

	port := ":" + os.Getenv("APP_PORT")
	e.Logger.Fatal(e.Start(port))
}
