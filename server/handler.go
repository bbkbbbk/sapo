package server

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultCookieExpires = 60

	//TODO: user real uid when integrate with LIFF
	uid = "123456789"
)

var (
	errorInvalidSpotifyAuthCode  = errors.New("invalid spotifyClient authorization code")
	errorInvalidSpotifyAuthState = errors.New("invalid spotifyClient auth state")
	errorUnableToGetCookie       = errors.New("unable to get cookie")
	errorUnableLogIn             = errors.New("unable to login to spotifyClient")
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
	events, err := h.botClient.ParseRequest(c.Request())
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			return c.JSON(http.StatusBadRequest, linebot.ErrInvalidSignature.Error())
		} else {
			return c.JSON(http.StatusInternalServerError, "[LINEMessageCallback]: unable to parse request")
		}
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if _, err = h.botClient.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
					logrus.Errorf("[LINEMessageCallback]: unable to reply message %v", err)
				}
			}
		}
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
