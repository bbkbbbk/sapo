package server

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultCookieExpires = 60
)

var (
	errorInvalidUID              = errors.New("invalid LINE user id")
	errorInvalidSpotifyAuthCode  = errors.New("invalid spotify authorization code")
	errorInvalidSpotifyAuthState = errors.New("invalid spotify auth state")
	errorUnableToGetCookie       = errors.New("unable to get cookie")
	errorUnableLogIn             = errors.New("unable to login to spotify")
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
	logrus.Error(err.Error())
	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}

func (h *Handler) HomePage(c echo.Context) error {
	return c.JSON(http.StatusOK, "Hello this is sapo")
}

func (h *Handler) LINECallback(c echo.Context) error {
	events, err := h.service.ParseLINERequest(c.Request())
	if err != nil {
		return h.returnError(err)
	}

	err = h.service.LINEEventsHandler(events)
	if err != nil {
		return h.returnError(err)
	}

	return c.JSON(http.StatusOK, "")
}

func (h *Handler) SignUp(c echo.Context) error {
	uid := c.QueryParam("uid")
	if uid == "" {
		return h.returnError(errorInvalidUID)
	}

	cookie := new(http.Cookie)
	cookie.Name = AuthState
	cookie.Value = uid
	cookie.Expires = time.Now().Add(defaultCookieExpires * time.Second)
	c.SetCookie(cookie)

	err := c.Redirect(302, h.service.GetSpotifyAuthURL(uid))
	if err != nil {
		return h.returnError(errors.Wrap(err, "[SignUp]: unable to redirect"))
	}

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

	uid := c.QueryParam("state")
	if uid == "" {
		return h.returnError(errorInvalidSpotifyAuthState)
	}

	storedState, err := c.Cookie(AuthState)
	if err != nil {
		return h.returnError(errorUnableToGetCookie)
	}

	if uid != storedState.Value {
		return h.returnError(errorInvalidSpotifyAuthState)
	}

	err = h.service.CreateAccount(uid, code)
	if err != nil {
		return h.returnError(errors.Wrap(err, "[SpotifyLoginCallback]: unable to create account"))
	}

	return c.JSON(http.StatusOK, "")
}
