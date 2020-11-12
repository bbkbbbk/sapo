package server

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultCookieExpires = 60
	uid                  = "123456789"
)

var (
	errorInvalidSpotifyAuthCode  = errors.New("invalid spotify authorization code")
	errorInvalidSpotifyAuthState = errors.New("invalid spotify auth state")
	errorUnableToGetCookie       = errors.New("unable to get cookie")
	errorUnableLogIn             = errors.New("unable to login to spotify")
)

type Handler struct {
	botClient *linebot.Client
	spotify   SpotifyService
}

func NewHandler(b *linebot.Client, s SpotifyService) Handler {
	return Handler{
		botClient: b,
		spotify:   s,
	}
}

func RandStringBytesMaskImprSrcSB(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6
		letterIdxMask = 1<<letterIdxBits - 1
		letterIdxMax  = 63 / letterIdxBits
	)

	var src = rand.NewSource(time.Now().UnixNano())
	sb := strings.Builder{}
	sb.Grow(n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func (h *Handler) returnError(err error) error {
	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}

func (h *Handler) HomePage(c echo.Context) error {
	return c.JSON(http.StatusOK, "Hello this is sapo")
}

func (h *Handler) PingCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, "[PingCheck]: ok")
}

func (h *Handler) LINEMessageCallback(c echo.Context) error {
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
	state := RandStringBytesMaskImprSrcSB(16)

	cookie := new(http.Cookie)
	cookie.Name = AuthState
	cookie.Value = state
	cookie.Expires = time.Now().Add(defaultCookieExpires * time.Second)
	c.SetCookie(cookie)

	c.Redirect(302, h.spotify.GetAuthURL(state))

	return nil
}

func (h *Handler) SpotifyLoginCallback(c echo.Context) error {
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

	err = h.spotify.GetTokenRequest(uid, code)
	if err != nil {
		return h.returnError(errors.Wrap(err, "[SpotifyLoginCallback]: unable to get token from spotify"))
	}

	return c.JSON(http.StatusOK, "")
}
