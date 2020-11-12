package server

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
)

type Service interface {
	CreateAccount(uid, code string) error
	GetSpotifyAuthURL(state string) string
	LINEEventsHandler(req *http.Request) error
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
	accToken, refToken, err := s.spotifyService.RequestToken(code)
	if err != nil {
		return errors.Wrap(err, "[s.CreateAccount]: unable to get token from spotify")
	}

	acc := Account{
		UID:          uid,
		AccessToken:  accToken,
		RefreshToken: refToken,
	}

	if _, err := s.repository.CreateAccount(acc); err != nil {
		return errors.Wrap(err, "[s.CreateAccount]: unable to create account")
	}

	return nil
}

func (s *service) LINEEventsHandler(req *http.Request) error {
	events, err := s.lineService.GetEvents(req)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			return errors.Wrapf(err, "[LINEEventsHandler]: invalid signature %v", linebot.ErrInvalidSignature.Error())
		}

		return errors.Wrapf(err, "[LINEEventsHandler]: unable to parse request")
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if err = s.lineService.TextMsgHandler(message.Text, event.ReplyToken); err != nil {
					return errors.Wrap(err, "[LINEEventsHandler]: unable to reply message %v")
				}
			}
		}
	}

	return nil
}
