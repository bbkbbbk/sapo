package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bbkbbbk/sapo/line"
	pkgMongo "github.com/bbkbbbk/sapo/pkg/mongo"
	"github.com/bbkbbbk/sapo/server"
	"github.com/bbkbbbk/sapo/spotify"
)

var (
	basedURL       string
	db             *mongo.Database
	lineService    line.Service
	richMenu       line.RichMenuMetadata
	spotifyService spotify.Service
)

func init() {
	basedURL = os.Getenv("BASED_URL_APP")
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

func init() {
	richMenu = line.RichMenuMetadata{
		Login:   os.Getenv("RICH_MENU_LOGIN"),
		Default: os.Getenv("RICH_MENU_DEFAULT"),
	}
}

func init() {
	lineService = line.NewLINEService(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
		richMenu,
	)
}

func init() {
	spotifyService = spotify.NewSpotifyService(
		os.Getenv("MY_CLIENT_ID"),
		os.Getenv("MY_CLIENT_SECRET"),
		basedURL,
	)
}

func main() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: strings.Split(os.Getenv("CORS_ALLOW_ORIGIN"), ","),
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodPut},
	}))

	repository := server.NewRepository(db)
	service := server.NewService(basedURL, lineService, spotifyService, repository)
	serverHandler := server.NewHandler(service)
	server.RoutesRegister(e, serverHandler)

	port := ":" + os.Getenv("APP_PORT")
	e.Logger.Fatal(e.Start(port))
}
