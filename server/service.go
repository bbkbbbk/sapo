package server

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
)

const (
	textEventSignUp = "signup"
	textEventEcho   = "echo"
)

type Service interface {
	CreateAccount(uid, code string) error
	GetSpotifyAuthURL(state string) string
	ParseLINERequest(req *http.Request) ([]*linebot.Event, error)
	LINEEventsHandler(events []*linebot.Event) error
	RandomString(n int) string
}

type service struct {
	lineService    LINEService
	spotifyService SpotifyService
	repository     Repository
}

func NewService(line LINEService, spotify SpotifyService, repo Repository) Service {
	return &service{
		lineService:    line,
		spotifyService: spotify,
		repository:     repo,
	}
}

func (s *service) RandomString(n int) string {
	const (
		letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
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

func (s *service) GetSpotifyAuthURL(state string) string {
	return s.spotifyService.GetAuthURL(state)
}

func (s *service) CreateAccount(uid, code string) error {
	now := time.Now()
	accToken, refToken, err := s.spotifyService.RequestToken(code)
	if err != nil {
		return errors.Wrap(err, "[s.CreateAccount]: unable to get token from spotify")
	}

	acc := Account{
		UID:          uid,
		AccessToken:  accToken,
		RefreshToken: refToken,
		CreatedAt:    &now,
	}

	if _, err := s.repository.CreateAccount(acc); err != nil {
		return errors.Wrap(err, "[s.CreateAccount]: unable to create account")
	}

	return nil
}

func (s *service) ParseLINERequest(req *http.Request) ([]*linebot.Event, error) {
	return s.lineService.ParseRequest(req)
}

func (s *service) LINEEventsHandler(events []*linebot.Event) error {
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			uid := event.Source.UserID

			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if err := s.textEventsHandler(uid, message.Text, event.ReplyToken); err != nil {
					return errors.Wrap(err, "[LINEEventsHandler]: unable to reply message")
				}
			}
		}
	}

	return nil
}

func (s *service) textEventsHandler(uid, msg, token string) error {
	switch msg {
	case textEventSignUp:
		if err := s.textEventSignUp(uid); err != nil {
			return errors.Wrap(err, "[TextEventsHandler]: unable to signup")
		}
	case textEventEcho:
		if err := s.lineService.EchoMsg(msg, token); err != nil {
			return errors.Wrap(err, "[TextEventsHandler]: unable to send echo message")
		}
	}

	return nil
}

func (s *service) textEventSignUp(uid string) error {
	url := fmt.Sprintf("https://sapo-wb87j.ondigitalocean.app/signup?uid=%s", uid)
	client := &http.Client{
		Timeout: time.Second * defaultTimeout,
	}
	res, err := client.Get(url)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			logrus.Warn("[TextEventsHandler]: unable to close response body", err)
		}
	}()
	if err != nil {
		return errors.Wrap(err, "[TextEventsHandler]: unable to signup")
	}

	return nil
}
