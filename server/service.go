package server

import (
	"net/http"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"

	"github.com/bbkbbbk/sapo/spotify"
)

const (
	defaultTimeout = 30

	textEventEcho = "echo"
)

type Service interface {
	CreateAccount(uid, code string) error
	GetSpotifyAuthURL(state string) string
	ParseLINERequest(req *http.Request) ([]*linebot.Event, error)
	LINEEventsHandler(events []*linebot.Event) error
	LINELinkUserToLoginRichMenu(uid string) error
	LINELinkUserToDefaultRichMenu(uid string) error
}

type service struct {
	basedURL       string
	lineService    LINEService
	spotifyService spotify.Service
	repository     Repository
}

func NewService(url string, line LINEService, spotify spotify.Service, repo Repository) Service {
	return &service{
		basedURL:       url,
		lineService:    line,
		spotifyService: spotify,
		repository:     repo,
	}
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

	profile, err := s.spotifyService.GetUserProfile(accToken)
	if err != nil {
		return errors.Wrap(err, "[s.CreateAccount]: unable to get spotify user profile")
	}
	spotifyId := profile.ID

	acc := Account{
		UID:          uid,
		SpotifyID:    spotifyId,
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
	case textEventEcho:
		if err := s.lineService.EchoMsg(msg, token); err != nil {
			return errors.Wrap(err, "[textEventsHandler]: unable to send echo message")
		}
	}

	return nil
}

func (s *service) LINELinkUserToLoginRichMenu(uid string) error {
	err := s.lineService.LinkUserToLoginRichMenu(uid)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) LINELinkUserToDefaultRichMenu(uid string) error {
	err := s.lineService.LinkUserToDefaultRichMenu(uid)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetAccountByUID(uid string) (*Account, error) {
	acc, err := s.repository.GetAccountByUID(uid)
	if err != nil {
		return nil, errors.Wrap(err, "[GetUserSpotifyToken]: unable to get user account token")
	}

	return acc, nil
}

func (s *service) CreateRecommendedPlaylistForUser(uid string) error {
	acc, err := s.GetAccountByUID(uid)
	if err != nil {
		return errors.Wrap(err, "[CreateRecommendedPlaylistForUser]: unable to get user account token")
	}
	spotifyId := acc.SpotifyID
	refreshToken := acc.RefreshToken

	accessToken, err := s.spotifyService.RequestAccessTokenFromRefreshToken(refreshToken)
	if err != nil {
		return errors.Wrap(err, "[CreateRecommendedPlaylistForUser]: unable to request access token")

	}

	err = s.spotifyService.CreateRecommendedPlaylistForUser(accessToken, spotifyId)
	if err != nil {
		return errors.Wrapf(err, "[CreateRecommendedPlaylistForUser]: unable to create recommended playlist for user id %s", uid)
	}

	return nil
}
