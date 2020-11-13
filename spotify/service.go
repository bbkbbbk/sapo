package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultTimeout = 30
	scopes         = "user-read-recently-played playlist-modify-public playlist-read-collaborative user-read-recently-played user-top-read user-library-read"

	AuthState         = "spotify-auth-state"
	LimitSeedSize     = 5
	LimitPlaylistSize = 25
)

var (
	errorUnableToGetToken = errors.New("unable to get token from spotify")
	errorInvalidSeed      = errors.New("invalid spotify seed")
)

type SpotifyService interface {
	GetAuthURL(state string) string
	RequestToken(code string) (string, string, error)
	RequestAccessTokenFromRefreshToken(token string) (string, error)
	GetUserProfile(token string) (*User, error)
	CreateRecommendedPlaylistForUser(token, uid string) error
}

type spotifyService struct {
	ClientID    string
	ClintSecret string
	CallbackURL string
}

type responseTokenBody struct {
	AccessToken    string `json:"access_token"`
	TokenType      string `json:"token_type"`
	Scope          string `json:"scope"`
	ExpirationTime int    `json:"expires_in"`
	RefreshToken   string `json:"refresh_token"`
}

func NewSpotifyService(id, secret, url string) SpotifyService {
	callbackURL := fmt.Sprintf("%s/spotify-callback", url)
	return &spotifyService{
		ClientID:    id,
		ClintSecret: secret,
		CallbackURL: callbackURL,
	}
}

func (s *spotifyService) makeRequest(req *http.Request) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Second * defaultTimeout,
	}

	res, err := client.Do(req)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			logrus.Warn("[makeRequest]: unable to close response body", err)
		}
	}()
	if err != nil {
		return nil, errors.Wrap(err, "[makeRequest]: unable to get response from spotify")
	}

	if res.StatusCode != http.StatusOK {
		logrus.Warn("[makeRequest]: unable to make a success response")
		return nil, errorUnableToGetToken
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "[makeRequest]: unable to read response body")
	}

	return body, nil
}

func (s *spotifyService) newAuthHeader() string {
	raw := fmt.Sprintf("%s:%s", s.ClientID, s.ClintSecret)
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))
	authHeader := fmt.Sprintf("Basic %s", encoded)

	return authHeader
}

func (s *spotifyService) newAuthAccessHeader(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}

func (s *spotifyService) GetAuthURL(state string) string {
	spotifyURL := "https://accounts.spotify.com/authorize"

	scope := url.QueryEscape(scopes)
	path := fmt.Sprintf("%s?client_id=%s&scope=%s&response_type=code&redirect_uri=%s&state=%s", spotifyURL, s.ClientID, scope, s.CallbackURL, state)

	return path
}

func (s *spotifyService) RequestToken(code string) (string, string, error) {
	spotifyURL := "https://accounts.spotify.com/api/token"

	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", s.CallbackURL)

	req, err := http.NewRequest("POST", spotifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", errors.Wrap(err, "[RequestToken]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthHeader())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

	res, err := s.makeRequest(req)
	if err != nil {
		return "", "", errors.Wrap(err, "[RequestToken]: unable to make request")
	}

	var token responseTokenBody
	err = json.Unmarshal(res, &token)
	if err != nil {
		return "", "", errors.Wrap(err, "[RequestToken]: unable to unmarshal response body")
	}

	accessToken := token.AccessToken
	refreshToken := token.RefreshToken

	return accessToken, refreshToken, nil
}

func (s *spotifyService) RequestAccessTokenFromRefreshToken(token string) (string, error) {
	spotifyURL := "https://accounts.spotify.com/api/token"

	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", token)

	req, err := http.NewRequest("POST", spotifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", errors.Wrap(err, "[RequestAccessTokenFromRefreshToken]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthHeader())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

	res, err := s.makeRequest(req)
	if err != nil {
		return "", errors.Wrap(err, "[RequestAccessTokenFromRefreshToken]: unable to make request")
	}

	var tokenRes responseTokenBody
	err = json.Unmarshal(res, &tokenRes)
	if err != nil {
		return "", errors.Wrap(err, "[RequestAccessTokenFromRefreshToken]: unable to unmarshal response body")
	}

	accessToken := tokenRes.AccessToken

	return accessToken, nil
}

func (s *spotifyService) GetSeeds(token string) ([]string, error) {
	spotifyURL := "https://api.spotify.com/v1/me/player/recently-played?limit=50"

	req, err := http.NewRequest("GET", spotifyURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "[GetSeeds]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthAccessHeader(token))

	res, err := s.makeRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "[GetSeeds]: unable to make request")
	}

	var playingHistory PlayHistoryObject
	err = json.Unmarshal(res, &playingHistory)
	if err != nil {
		return nil, errors.Wrap(err, "[GetSeeds]: unable to unmarshal response body")
	}

	seedTracks := []string{}
	for i, item := range playingHistory.Items {
		if i%10 == 0 {
			seedTrack := item.Track.ID
			seedTracks = append(seedTracks, seedTrack)
		}
	}

	return seedTracks, nil
}

func (s *spotifyService) GetRecommendationsBasedOnSeeds(token string, seeds []string) ([]string, error) {
	if len(seeds) > LimitSeedSize {
		return nil, errorInvalidSeed
	}

	spotifyURL := "https://api.spotify.com/v1/recommendations"
	limit := LimitPlaylistSize
	seedTracks := strings.Join(seeds, ",")

	path := fmt.Sprintf("%s?limit=%d&seed_tracks=%s", spotifyURL, limit, seedTracks)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "[GetRecommendationsBasedOnSeeds]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthAccessHeader(token))

	res, err := s.makeRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "[GetRecommendationsBasedOnSeeds]: unable to make request")
	}

	var tracks SimplifiedTracks
	err = json.Unmarshal(res, &tracks)
	if err != nil {
		return nil, errors.Wrap(err, "[GetRecommendationsBasedOnSeeds]: unable to unmarshal response body")
	}

	uris := []string{}
	for _, track := range tracks.Tracks {
		uri := track.URI
		uris = append(uris, uri)
	}

	return uris, nil
}

func (s *spotifyService) CreatePlaylistForUser(token, uid string) (string, error) {
	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", uid)
	name := fmt.Sprintf("Tracks for you %s", time.Now().Format("2000-05-19"))

	form := url.Values{}
	form.Add("name", name)
	form.Add("description", "Recommended tracks created by sapo")

	req, err := http.NewRequest("POST", spotifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", errors.Wrap(err, "[CreatePlaylistForUser]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthAccessHeader(token))
	req.Header.Add("Content-Type", "application/json")

	res, err := s.makeRequest(req)
	if err != nil {
		return "", errors.Wrap(err, "[CreatePlaylistForUser]: unable to make request")
	}

	var playlist SimplifiedObject
	err = json.Unmarshal(res, &playlist)
	if err != nil {
		return "", errors.Wrap(err, "[CreatePlaylistForUser]: unable to unmarshal response body")
	}

	id := playlist.ID

	return id, nil
}

func (s *spotifyService) AddTracksToPlaylist(token, id string, uris []string) error {
	urisParam := strings.Join(uris, ",")
	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?uris=%s", id, urisParam)

	req, err := http.NewRequest("POST", spotifyURL, nil)
	if err != nil {
		return errors.Wrap(err, "[AddTracksToPlaylist]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthAccessHeader(token))
	req.Header.Add("Content-Type", "application/json")

	_, err = s.makeRequest(req)
	if err != nil {
		return errors.Wrap(err, "[AddTracksToPlaylist]: unable to make request")
	}

	return nil
}

func (s *spotifyService) CreateRecommendedPlaylistForUser(token, uid string) error {
	accessToken, err := s.RequestAccessTokenFromRefreshToken(token)
	if err != nil {
		return errors.Wrap(err, "[CreateRecommendedPlaylistForUser]: unable to request access token")
	}

	seeds, err := s.GetSeeds(accessToken)
	if err != nil {
		return errors.Wrap(err, "[CreateRecommendedPlaylistForUser]: unable to get seeds")
	}

	uris, err := s.GetRecommendationsBasedOnSeeds(accessToken, seeds)
	if err != nil {
		return errors.Wrap(err, "[CreateRecommendedPlaylistForUser]: unable to get uris from seeds")
	}

	playlistId, err := s.CreatePlaylistForUser(token, uid)
	if err != nil {
		return errors.Wrap(err, "[CreateRecommendedPlaylistForUser]: unable to create playlist")
	}

	err = s.AddTracksToPlaylist(token, playlistId, uris)
	if err != nil {
		return errors.Wrap(err, "[CreateRecommendedPlaylistForUser]: unable to add track to a playlist")
	}

	return nil
}

func (s *spotifyService) GetUserProfile(token string) (*User, error) {
	spotifyURL := "https://api.spotify.com/v1/me"

	req, err := http.NewRequest("GET", spotifyURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "[GetUserProfile]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthAccessHeader(token))

	res, err := s.makeRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "[GetUserProfile]: unable to make request")
	}

	var user User
	err = json.Unmarshal(res, &user)
	if err != nil {
		return nil, errors.Wrap(err, "[GetUserProfile]: unable to unmarshal response body")
	}

	return &user, nil
}
