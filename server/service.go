package server

import (
	"net/http"
	"strconv"
	"strings"
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
	textEventMyTopTracks = "my top tracks"
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

		flex := s.createPlaylistFlexMsg(playlist)

		if err := s.lineService.ReplyFlexMsg(token, *flex); err != nil {
			return errors.Wrap(err, "[textEventsHandler]: unable to send flex message")
		}
	case textEventMyTopTracks:
		tracks, albums, err := s.GetTopTracksWithAlbums(uid)
		if err != nil {
			return errors.Wrapf(err, "[textEventsHandler]: unable to get top tracks for user id %s", uid)
		}

		flex := s.CreateTopTracksFlexMsg(tracks, albums)

		if err := s.lineService.ReplyFlexMsg(token, *flex); err != nil {
			return errors.Wrap(err, "[textEventsHandler]: unable to send flex message")
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

	playlist, err := s.spotifyService.GetPlaylist(accessToken, playlistId)
	if err != nil {
		return nil, errors.Wrap(err, "[createRecommendedPlaylistForUser]: unable to playlist detail")
	}

	return playlist, nil
}

func (s *service) createPlaylistFlexMsg(playlist *spotify.Playlist) *message.Flex {
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

	return &flex
}

func (s *service) GetTopTracksWithAlbums(uid string) ([]*spotify.Track, []*spotify.Album, error) {
	acc, err := s.getAccountByUID(uid)
	if err != nil {
		return nil, nil, errors.Wrap(err, "[GetTopTracksWithAlbums]: unable to get user profile")
	}
	refreshToken := acc.RefreshToken

	accessToken, err := s.spotifyService.RequestAccessTokenFromRefreshToken(refreshToken)
	if err != nil {
		return nil, nil, errors.Wrap(err, "[GetTopTracksWithAlbums]: unable to request access token")
	}

	tracks, err := s.spotifyService.GetTopTracks(accessToken)
	if err != nil {
		return nil, nil, errors.Wrap(err, "[GetTopTracksWithAlbums]: unable to get user's top tracks")
	}

	albumIDs := s.findUniqueAlbumIDsFromTracks(tracks)

	albums, err := s.spotifyService.GetAlbums(accessToken, albumIDs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "[GetTopTracksWithAlbums]: unable to albums from ids")
	}

	return tracks, albums, nil
}

func (s *service) CreateTopTracksFlexMsg(tracks []*spotify.Track, albums []*spotify.Album) *message.Flex {
	AlbumIDMapImageURL := map[string]string{}
	for _, album := range albums {
		AlbumIDMapImageURL[album.ID] = album.Images[0].URL
	}

	boxes := []message.BoxWithImage{}
	for _, track := range tracks {
		artists := []string{}
		for _, a := range track.Artists {
			artists = append(artists, a.Name)
		}

		box := message.BoxWithImage {
			Header: track.Name,
			Text: strings.Join(artists, ", "),
			LeftText: strconv.Itoa(track.Duration),
			ImageURL: AlbumIDMapImageURL[track.Album.ID],
			URL: track.ExternalURLs.URL,
		}
		boxes = append(boxes, box)
	}

	now := time.Now()
	flex := message.NewBubbleReceipt(
		"My Top Tracks",
		"sapo",
		"My Top Tracks",
		now.Format("02 January 2006"),
		boxes,
		)

	return &flex
}

func (s *service) findUniqueAlbumIDsFromTracks(tracks []*spotify.Track)[]string {
	albums := map[string]string{}
	for _, track := range tracks {
		name := track.Album.Name
		id := track.Album.ID
		albums[id] = name
	}

	ids := []string{}
	for _, id := range albums {
		ids = append(ids, id)
	}

	return ids
}