package server

import (
	"net/http"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type LINEService interface {
	GetEvents(req *http.Request) ([]*linebot.Event, error)
	TextMsgHandler(msg, token string) error
}

type lineService struct {
	lineClient *linebot.Client
}

func NewLINEService(secret, token string) LINEService {
	bot, err := linebot.New(secret, token)
	if err != nil {
		logrus.Warnf("[init]: unable to initialize line line client %v", err)
	}

	return &lineService{
		lineClient: bot,
	}
}

func (l *lineService) GetEvents(req *http.Request) ([]*linebot.Event, error) {
	events, err := l.lineClient.ParseRequest(req)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (l *lineService) TextMsgHandler(msg, token string) error {
	replyMsg := l.echoMsg(msg)

	_, err := l.lineClient.ReplyMessage(token, replyMsg).Do()
	if err != nil {
		return errors.Wrap(err, "[TextMsgHandler]: unable to send a reply text message")
	}

	return nil
}

func (l *lineService) echoMsg(msg string) *linebot.TextMessage {
	return linebot.NewTextMessage(msg)
}
