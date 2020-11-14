package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type LINEService interface {
	ParseRequest(req *http.Request) ([]*linebot.Event, error)
	SendMessage(msg, token string) error
	LinkUserToLoginRichMenu(uid string) error
	LinkUserToDefaultRichMenu(uid string) error
}

type lineService struct {
	lineClient *linebot.Client
	richMenu   LINERichMenuMetadata
	token      string
}

type LINERichMenuMetadata struct {
	Login   string
	Default string
}

func NewLINEService(secret, token string, menu LINERichMenuMetadata) LINEService {
	bot, err := linebot.New(secret, token)
	if err != nil {
		logrus.Warnf("[NewLINEService]: unable to initialize line line client %v", err)
	}

	return &lineService{
		lineClient: bot,
		token:      token,
		richMenu:   menu,
	}
}

func (l *lineService) newAuthHeader() string {
	return fmt.Sprintf("Bearer %s", l.token)
}

func (l *lineService) ParseRequest(req *http.Request) ([]*linebot.Event, error) {
	return l.lineClient.ParseRequest(req)
}

func (l *lineService) SendMessage(msg, token string) error {
	replyMsg := linebot.NewTextMessage(msg)
	_, err := l.lineClient.ReplyMessage(token, replyMsg).Do()

	if err != nil {
		return errors.Wrap(err, "[EchoMsg]: unable to send a reply text message")
	}

	return nil
}

func (l *lineService) LinkUserToLoginRichMenu(uid string) error {
	rid := l.richMenu.Login
	err := l.linkUserToRichMenu(uid, rid)
	if err != nil {
		return err
	}

	return nil
}

func (l *lineService) LinkUserToDefaultRichMenu(uid string) error {
	rid := l.richMenu.Default
	err := l.linkUserToRichMenu(uid, rid)
	if err != nil {
		return err
	}

	return nil
}

func (l *lineService) linkUserToRichMenu(uid, rid string) error {
	lineURL := fmt.Sprintf("https://api.line.me/v2/bot/user/%s/richmenu/%s", uid, rid)

	req, err := http.NewRequest("POST", lineURL, nil)
	if err != nil {
		return errors.Wrap(err, "[linkUserToRichMenu]: unable to create request")
	}

	req.Header.Add("Authorization", l.newAuthHeader())

	client := &http.Client{
		Timeout: time.Second * defaultTimeout,
	}
	res, err := client.Do(req)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			logrus.Warn("[linkUserToRichMenu]: unable to close response body", err)
		}
	}()
	if err != nil {
		return errors.Wrapf(err, "[linkUserToRichMenu]: unable to request rich menu change for user id %s and rich menu id %s", uid, rid)
	}

	return nil
}
