package line

import (
	"fmt"
	"net/http"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultTimeout = 30
)

type Service interface {
	ParseRequest(req *http.Request) ([]*linebot.Event, error)
	SendMessage(msg, token string) error
	LinkUserToLoginRichMenu(uid string) error
	LinkUserToDefaultRichMenu(uid string) error
}

type service struct {
	lineClient *linebot.Client
	richMenu   RichMenuMetadata
	token      string
}

type RichMenuMetadata struct {
	Login   string
	Default string
}

func NewLINEService(secret, token string, menu RichMenuMetadata) Service {
	bot, err := linebot.New(secret, token)
	if err != nil {
		logrus.Warnf("[NewLINEService]: unable to initialize line line client %v", err)
	}

	return &service{
		lineClient: bot,
		token:      token,
		richMenu:   menu,
	}
}

func (s *service) newAuthHeader() string {
	return fmt.Sprintf("Bearer %s", s.token)
}

func (s *service) ParseRequest(req *http.Request) ([]*linebot.Event, error) {
	return s.lineClient.ParseRequest(req)
}

func (s *service) SendMessage(msg, token string) error {
	replyMsg := linebot.NewTextMessage(msg)
	_, err := s.lineClient.ReplyMessage(token, replyMsg).Do()

	if err != nil {
		return errors.Wrap(err, "[EchoMsg]: unable to send a reply text message")
	}

	return nil
}

func (s *service) LinkUserToLoginRichMenu(uid string) error {
	rid := s.richMenu.Login
	err := s.linkUserToRichMenu(uid, rid)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) LinkUserToDefaultRichMenu(uid string) error {
	rid := s.richMenu.Default
	err := s.linkUserToRichMenu(uid, rid)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) linkUserToRichMenu(uid, rid string) error {
	lineURL := fmt.Sprintf("https://api.line.me/v2/bot/user/%s/richmenu/%s", uid, rid)

	req, err := http.NewRequest("POST", lineURL, nil)
	if err != nil {
		return errors.Wrap(err, "[linkUserToRichMenu]: unable to create request")
	}

	req.Header.Add("Authorization", s.newAuthHeader())

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
