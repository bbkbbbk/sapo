package server

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
)

type Handler struct {
	botClient *linebot.Client
}

func NewHandler(b *linebot.Client) Handler {
	return Handler{
		botClient: b,
	}
}

func (h *Handler) returnError(err error) error {
	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}

func (h *Handler) PingCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, "[PingCheck]: ok")
}

func (h *Handler) Callback(c echo.Context) error {
	events, err := h.botClient.ParseRequest(c.Request())
	if err != nil {
		return h.returnError(errors.Wrap(err, "[Callback]: unable to parse request"))
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if _, err = h.botClient.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}

	return c.JSON(200, "")
}
