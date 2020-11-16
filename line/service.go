package line

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/bbkbbbk/sapo/line/message"
)

const (
	defaultTimeout = 30
)

type Service interface {
	ParseRequest(req *http.Request) ([]*linebot.Event, error)
	SendTextMessage(msg, token string) error
	SendTextMessageWithQuickReplies(token, msg string, quickReplies *linebot.QuickReplyItems) error
	SendFlexMessage(token string, msg *linebot.FlexMessage) error
	LinkUserToLoginRichMenu(uid string) error
	LinkUserToDefaultRichMenu(uid string) error
	ReplyFlexMsg(replyToken string, flex message.Flex) error
	PushFlexMsg(uid string, flex message.Flex) error
}

type service struct {
	lineClient   *linebot.Client
	richMenu     RichMenuMetadata
	channelToken string
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
		lineClient:   bot,
		channelToken: token,
		richMenu:     menu,
	}
}

func (s *service) newAuthHeader() string {
	return fmt.Sprintf("Bearer %s", s.channelToken)
}

func (s *service) ParseRequest(req *http.Request) ([]*linebot.Event, error) {
	return s.lineClient.ParseRequest(req)
}

func (s *service) SendTextMessage(token, msg string) error {
	replyMsg := linebot.NewTextMessage(msg)
	_, err := s.lineClient.ReplyMessage(token, replyMsg).Do()
	if err != nil {
		return errors.Wrap(err, "[SendTextMessage]: unable to send a reply text message")
	}

	return nil
}

func (s *service) SendTextMessageWithQuickReplies(token, msg string, quickReplies *linebot.QuickReplyItems) error {
	replyMsg := linebot.NewTextMessage(msg).WithQuickReplies(quickReplies)
	_, err := s.lineClient.ReplyMessage(token, replyMsg).Do()
	if err != nil {
		return errors.Wrap(err, "[SendTextMessageWithQuickReplies]: unable to send a reply text message")
	}

	return nil
}

func (s *service) SendFlexMessage(token string, msg *linebot.FlexMessage) error {
	_, err := s.lineClient.ReplyMessage(token, msg).Do()
	if err != nil {
		return errors.Wrap(err, "[SendFlexMessage]: unable to send a reply flex message")
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

func (s *service) ReplyFlexMsg(replyToken string, flex message.Flex) error {
	lineURL := "https://api.line.me/v2/bot/message/reply"

	msg := message.Reply{
		ReplyToken: replyToken,
		Message:    flex,
	}

	req, err := http.NewRequest("POST", lineURL, bytes.NewBuffer(msg.ToJson()))
	if err != nil {
		return errors.Wrap(err, "[SendReplyFlexMsg]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthHeader())
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * defaultTimeout,
	}
	res, err := client.Do(req)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			logrus.Warn("[SendReplyFlexMsg]: unable to close response body", err)
		}
	}()
	if err != nil {
		return errors.Wrap(err, "[SendReplyFlexMsg]: unable to make a success request")
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		logrus.Info(res.StatusCode)
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logrus.Warn(err)
		}

		var b interface{}
		err = json.Unmarshal(body, &b)
		if err != nil {
			logrus.Warn(err)
		}
		logrus.Info(b)
		logrus.Info(flex.ToFlex())
	}

	return nil
}

func (s *service) PushFlexMsg(uid string, flex message.Flex) error {
	lineURL := "https://api.line.me/v2/bot/message/push"

	msg := message.Push{
		ToID:    uid,
		Message: flex,
	}

	req, err := http.NewRequest("POST", lineURL, bytes.NewBuffer(msg.ToJson()))
	if err != nil {
		return errors.Wrap(err, "[SendReplyFlexMsg]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthHeader())
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * defaultTimeout,
	}
	res, err := client.Do(req)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			logrus.Warn("[SendReplyFlexMsg]: unable to close response body", err)
		}
	}()
	if err != nil {
		return errors.Wrap(err, "[SendReplyFlexMsg]: unable to make a success request")
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		logrus.Info(res.StatusCode)
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logrus.Warn(err)
		}

		var b interface{}
		err = json.Unmarshal(body, &b)
		if err != nil {
			logrus.Warn(err)
		}
		logrus.Info(b)
		logrus.Info(flex.ToFlex())
	}

	return nil
}
