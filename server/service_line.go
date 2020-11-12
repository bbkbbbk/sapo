package server

import (
	"net/http"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type LINEService interface {
	ParseRequest(req *http.Request) ([]*linebot.Event, error)
	EchoMsg(msg, token string) error
}

type lineService struct {
	lineClient *linebot.Client
}

func NewLINEService(secret, token string) LINEService {
	bot, err := linebot.New(secret, token)
	if err != nil {
		logrus.Warnf("[NewLINEService]: unable to initialize line line client %v", err)
	}

	return &lineService{
		lineClient: bot,
	}
}

func (l *lineService) ParseRequest(req *http.Request) ([]*linebot.Event, error) {
	return l.lineClient.ParseRequest(req)
}

func (l *lineService) EchoMsg(msg, token string) error {
	replyMsg := linebot.NewTextMessage(msg)
	_, err := l.lineClient.ReplyMessage(token, replyMsg).Do()

	if err != nil {
		return errors.Wrap(err, "[TextMsgHandler]: unable to send a reply text message")
	}

	return nil
}
