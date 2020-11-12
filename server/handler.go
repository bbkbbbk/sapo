package server

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

const (
	defaultCookieExpires = 60

	//TODO: user real uid when integrate with LIFF
	uid = "123456789"
)

var (
	errorInvalidSpotifyAuthCode  = errors.New("invalid spotifyService authorization code")
	errorInvalidSpotifyAuthState = errors.New("invalid spotifyService auth state")
	errorUnableToGetCookie       = errors.New("unable to get cookie")
	errorUnableLogIn             = errors.New("unable to login to spotifyService")
)

type Handler struct {
	service Service
}

func NewHandler(s Service) Handler {
	return Handler{
		service: s,
	}
}

func (h *Handler) returnError(err error) error {
	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}

func (h *Handler) HomePage(c echo.Context) error {
	return c.JSON(http.StatusOK, "Hello this is sapo")
}

func (h *Handler) LINECallback(c echo.Context) error {
	err := h.service.LINEEventsHandler(c.Request())
	if err != nil {
		return h.returnError(err)
	}

	return c.JSON(http.StatusOK, "")
}

func (h *Handler) SignUp(c echo.Context) error {
	state := h.service.RandomString(16)

	cookie := new(http.Cookie)
	cookie.Name = AuthState
	cookie.Value = state
	cookie.Expires = time.Now().Add(defaultCookieExpires * time.Second)
	c.SetCookie(cookie)

	c.Redirect(302, h.service.GetSpotifyAuthURL(state))

	return c.JSON(http.StatusOK, "")
}

func (h *Handler) SpotifyCallback(c echo.Context) error {
	errParam := c.QueryParam("error")
	if errParam != "" {
		return h.returnError(errors.Wrapf(errorUnableLogIn, "[SpotifyLoginCallback]: unable to login to spotify due to %v", errParam))
	}

	code := c.QueryParam("code")
	if code == "" {
		return h.returnError(errorInvalidSpotifyAuthCode)
	}

	state := c.QueryParam("state")
	if state == "" {
		return h.returnError(errorInvalidSpotifyAuthState)
	}

	storedState, err := c.Cookie(AuthState)
	if err != nil {
		return h.returnError(errorUnableToGetCookie)
	}

	if state != storedState.Value {
		return h.returnError(errorInvalidSpotifyAuthState)
	}

	err = h.service.CreateAccount(uid, code)
	if err != nil {
		return h.returnError(errors.Wrap(err, "[SpotifyLoginCallback]: unable to create account"))
	}

	return c.JSON(http.StatusOK, "")
}
