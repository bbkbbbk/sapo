package server

import (
	"net/http"
	"time"

	"github.com/bbkbbbk/sapo/line"
	"github.com/bbkbbbk/sapo/line/message"
	"github.com/bbkbbbk/sapo/spotify"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
)

const (
	defaultTimeout   = 30
	defaultFlexColor = "373C41CC"

	textEventEcho           = "echo"
	textEventCreatePlaylist = "create playlist"
)

type Service interface {
	CreateAccount(uid, code string) error
	GetSpotifyAuthURL(state string) string
	ParseLINERequest(req *http.Request) ([]*linebot.Event, error)
	LINEEventsHandler(events []*linebot.Event) error
	LINELinkUserToLoginRichMenu(uid string) error
	LINELinkUserToDefaultRichMenu(uid string) error
	TestFlex(uid string) error
}

type service struct {
	basedURL       string
	lineService    line.Service
	spotifyService spotify.Service
	repository     Repository
}

func NewService(url string, line line.Service, spotify spotify.Service, repo Repository) Service {
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

func (s *service) textEventsHandler(uid, msg, token string) error {
	switch msg {
	case textEventEcho:
		if err := s.lineService.SendTextMessage(token, msg); err != nil {
			return errors.Wrap(err, "[textEventsHandler]: unable to send message")
		}
	case textEventCreatePlaylist:
		playlist, err := s.createRecommendedPlaylistForUser(uid)
		if err != nil {
			return errors.Wrapf(err, "[textEventsHandler]: unable to create recommended playlist to user id %s", uid)
		}

		flex, err := s.createPlaylistFlexMsg(playlist)
		if err != nil {
			return errors.Wrapf(err, "[textEventsHandler]: unable to create playlist flex message")
		}

		if err := s.lineService.ReplyFlexMsg(token, *flex); err != nil {
			return errors.Wrap(err, "[textEventsHandler]: unable to send message")
		}
	}

	return nil
}

func (s *service) getAccountByUID(uid string) (*Account, error) {
	acc, err := s.repository.GetAccountByUID(uid)
	if err != nil {
		return nil, errors.Wrap(err, "[GetAccountByUID]: unable to get user account token")
	}

	return acc, nil
}

func (s *service) createRecommendedPlaylistForUser(uid string) (*spotify.Playlist, error) {
	acc, err := s.getAccountByUID(uid)
	if err != nil {
		return nil, errors.Wrap(err, "[createRecommendedPlaylistForUser]: unable to get user profile")
	}
	spotifyId := acc.SpotifyID
	refreshToken := acc.RefreshToken

	accessToken, err := s.spotifyService.RequestAccessTokenFromRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.Wrap(err, "[createRecommendedPlaylistForUser]: unable to request access token")
	}

	playlistId, err := s.spotifyService.CreateRecommendedPlaylistForUser(accessToken, spotifyId)
	if err != nil {
		return nil, errors.Wrap(err, "[createRecommendedPlaylistForUser]: unable to create playlist")
	}

	playlist, err := s.spotifyService.GetPlaylistByID(accessToken, playlistId)
	if err != nil {
		return nil, errors.Wrap(err, "[createRecommendedPlaylistForUser]: unable to playlist detail")
	}

	return playlist, nil
}

func (s *service) createPlaylistFlexMsg(playlist *spotify.Playlist) (*message.Flex, error) {
	altText := "Playlist for you"
	buttonLabel := "go to playlist"

	flex := message.NewBubbleWithButton(
		altText,
		playlist.Name,
		playlist.Description,
		buttonLabel,
		playlist.ExternalURLs.URL,
		playlist.Images[0].URL,
		defaultFlexColor,
	)

	return &flex, nil
}

func (s *service) TestFlex(uid string) error {
	flex := message.NewBubbleWithButton(
		"test",
		"https://mosaic.scdn.co/640/ab67616d0000b2730cdb4b03fd27a1301592a5e3ab67616d0000b2733ceefb49194de6fcd2ffe4c7ab67616d0000b27347669a9be7d201ea97bbd3eeab67616d0000b2736029febcb938b2b8cb279b47",
		"track sth",
		"tracks by sapo",
		"go",
		"https://open.spotify.com/playlist/0O68SXXRKfcHt5iJkWQ7xD",
		defaultFlexColor,
	)

	if err := s.lineService.ReplyFlexMsg("8197f23b7e454f22a081d0b2b37415d4", flex); err != nil {
		return err
	}

	return nil
}
