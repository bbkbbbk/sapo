package server

import "github.com/labstack/echo"

func RoutesRegister(e *echo.Echo, h Handler) {
	e.GET("/", h.HomePage)
	e.POST("/line-callback", h.LINECallback)
	e.GET("/signup", h.SignUp)
	e.GET("/spotify-callback", h.SpotifyCallback)
}
