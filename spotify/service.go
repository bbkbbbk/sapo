package spotify

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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
	errorInvalidSeed = errors.New("invalid spotify seed")
)

type Service interface {
	GetAuthURL(state string) string
	RequestToken(code string) (string, string, error)
	RequestAccessTokenFromRefreshToken(token string) (string, error)
	GetUserProfile(token string) (*User, error)
	GetPlaylist(token, id string) (*Playlist, error)
	GetAlbums(token string, ids []string) ([]*Album, error)
	GetTopArtists(token string) ([]*Artist, error)
	GetTopTracks(token string) ([]*Track, error)
	CreateRecommendedPlaylistForUser(token, uid string) (string, error)
}

type service struct {
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

type requestCreatePlaylist struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewSpotifyService(id, secret, url string) Service {
	callbackURL := fmt.Sprintf("%s/spotify-callback", url)
	return &service{
		ClientID:    id,
		ClintSecret: secret,
		CallbackURL: callbackURL,
	}
}

func (s *service) makeAuthRequest(form url.Values) ([]byte, error) {
	spotifyURL := "https://accounts.spotify.com/api/token"

	req, err := http.NewRequest("POST", spotifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "[makeAuthRequest]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthHeader())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

	client := &http.Client{
		Timeout: time.Second * defaultTimeout,
	}

	res, err := client.Do(req)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			logrus.Warn("[makeAuthRequest]: unable to close response body", err)
		}
	}()
	if err != nil {
		return nil, errors.Wrap(err, "[makeAuthRequest]: unable to get response from spotify")
	}

	if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
		return nil, errors.Wrap(err, "[makeAuthRequest]: unable to make a success request")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "[makeAuthRequest]: unable to read response body")
	}

	return body, nil
}

func (s *service) makeRequest(token, method, url string, reqBody io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "[makeRequest]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthAccessHeader(token))
	req.Header.Add("Content-Type", "application/json")

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

	if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
		return nil, errors.Wrap(err, "[makeRequest]: unable to make a success request")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "[makeRequest]: unable to read response body")
	}

	return body, nil
}

func (s *service) newAuthHeader() string {
	raw := fmt.Sprintf("%s:%s", s.ClientID, s.ClintSecret)
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))
	authHeader := fmt.Sprintf("Basic %s", encoded)

	return authHeader
}

func (s *service) newAuthAccessHeader(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}

func (s *service) GetAuthURL(state string) string {
	spotifyURL := "https://accounts.spotify.com/authorize"

	scope := url.QueryEscape(scopes)
	path := fmt.Sprintf("%s?client_id=%s&scope=%s&response_type=code&redirect_uri=%s&state=%s", spotifyURL, s.ClientID, scope, s.CallbackURL, state)

	return path
}

func (s *service) RequestToken(code string) (string, string, error) {
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", s.CallbackURL)

	res, err := s.makeAuthRequest(form)
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

func (s *service) RequestAccessTokenFromRefreshToken(token string) (string, error) {
	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", token)

	res, err := s.makeAuthRequest(form)
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

func (s *service) GetCurrentTrackSeeds(token string) ([]string, error) {
	spotifyURL := "https://api.spotify.com/v1/me/player/recently-played?limit=50"

	res, err := s.makeRequest(token, http.MethodGet, spotifyURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "[GetSeeds]: unable to make request")
	}

	var obj Paging
	err = json.Unmarshal(res, &obj)
	if err != nil {
		return nil, errors.Wrap(err, "[GetSeeds]: unable to unmarshal response body")
	}

	seedTracks := []string{}
	for i, item := range obj.Items {
		history := item.(PlayingHistory)
		if i%10 == 0 {
			seedTrack := history.Track.ID
			seedTracks = append(seedTracks, seedTrack)
		}
	}

	return seedTracks, nil
}

func (s *service) GetRecommendationsBasedOnSeeds(token string, seeds []string) ([]string, error) {
	if len(seeds) > LimitSeedSize {
		return nil, errorInvalidSeed
	}

	spotifyURL := "https://api.spotify.com/v1/recommendations"
	limit := LimitPlaylistSize
	seedTracks := strings.Join(seeds, ",")

	path := fmt.Sprintf("%s?limit=%d&seed_tracks=%s", spotifyURL, limit, seedTracks)

	res, err := s.makeRequest(token, http.MethodGet, path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "[GetRecommendationsBasedOnSeeds]: unable to make request")
	}

	var tracks Tracks
	err = json.Unmarshal(res, &tracks)
	if err != nil {
		return nil, errors.Wrap(err, "[GetRecommendationsBasedOnSeeds]: unable to unmarshal response body")
	}

	uris := []string{}
	for _, track := range tracks.Items {
		uri := track.URI
		uris = append(uris, uri)
	}

	return uris, nil
}

func (s *service) CreatePlaylistForUser(token, uid string) (string, error) {
	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", uid)
	now := time.Now()
	name := fmt.Sprintf("%s Tracks for you", now.Format("2006-01-02"))

	reqCreate := requestCreatePlaylist{
		Name:        name,
		Description: "Playlist created by sapo",
	}
	body, err := json.Marshal(&reqCreate)
	if err != nil {
		return "", errors.Wrap(err, "[CreatePlaylistForUser]: unable to marshal request body")
	}

	res, err := s.makeRequest(token, http.MethodPost, spotifyURL, bytes.NewReader(body))
	if err != nil {
		return "", errors.Wrap(err, "[CreatePlaylistForUser]: unable to make request")
	}

	var playlist Playlist
	err = json.Unmarshal(res, &playlist)
	if err != nil {
		return "", errors.Wrap(err, "[CreatePlaylistForUser]: unable to unmarshal response body")
	}

	id := playlist.ID

	return id, nil
}

func (s *service) AddTracksToPlaylist(token, id string, uris []string) error {
	urisParam := strings.Join(uris, ",")
	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?uris=%s", id, urisParam)

	_, err := s.makeRequest(token, http.MethodPost, spotifyURL, nil)
	if err != nil {
		return errors.Wrap(err, "[AddTracksToPlaylist]: unable to make request")
	}

	return nil
}

func (s *service) CreateRecommendedPlaylistForUser(token, uid string) (string, error) {
	seeds, err := s.GetCurrentTrackSeeds(token)
	if err != nil {
		return "", errors.Wrap(err, "[CreateRecommendedPlaylistForUser]: unable to get seeds")
	}

	uris, err := s.GetRecommendationsBasedOnSeeds(token, seeds)
	if err != nil {
		return "", errors.Wrap(err, "[CreateRecommendedPlaylistForUser]: unable to get uris from seeds")
	}

	playlistId, err := s.CreatePlaylistForUser(token, uid)
	if err != nil {
		return "", errors.Wrap(err, "[CreateRecommendedPlaylistForUser]: unable to create playlist")
	}

	err = s.AddTracksToPlaylist(token, playlistId, uris)
	if err != nil {
		return "", errors.Wrap(err, "[CreateRecommendedPlaylistForUser]: unable to add track to a playlist")
	}

	return playlistId, nil
}

func (s *service) GetUserProfile(token string) (*User, error) {
	spotifyURL := "https://api.spotify.com/v1/me"

	res, err := s.makeRequest(token, http.MethodGet, spotifyURL, nil)
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

func (s *service) GetPlaylist(token, id string) (*Playlist, error) {
	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s", id)

	res, err := s.makeRequest(token, http.MethodGet, spotifyURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "[GetPlaylist]: unable to make request")
	}

	var playlist Playlist
	err = json.Unmarshal(res, &playlist)
	if err != nil {
		return nil, errors.Wrap(err, "[GetPlaylist]: unable to unmarshal response body")
	}

	return &playlist, nil
}

func (s *service) GetTopArtists(token string) ([]*Artist, error) {
	spotifyURL := "https://api.spotify.com/v1/me/top/artists?limit=5&time_range=short_term"

	res, err := s.makeRequest(token, http.MethodGet, spotifyURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "[GetUserTopArtists]: unable to make request")
	}

	var obj Paging
	err = json.Unmarshal(res, &obj)
	if err != nil {
		return nil, errors.Wrap(err, "[GetUserTopArtists]: unable to unmarshal response body")
	}

	artists := []*Artist{}
	for _, item := range obj.Items {
		artist := item.(Artist)
		artists = append(artists, &artist)
	}

	return artists, nil
}

func (s *service) GetTopTracks(token string) ([]*Track, error) {
	spotifyURL := "https://api.spotify.com/v1/me/top/tracks?limit=5&time_range=short_term"

	res, err := s.makeRequest(token, http.MethodGet, spotifyURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "[GetUserTopTracks]: unable to make request")
	}

	var obj Paging
	err = json.Unmarshal(res, &obj)
	if err != nil {
		return nil, errors.Wrap(err, "[GetUserTopTracks]: unable to unmarshal response body")
	}

	tracks := []*Track{}
	for _, item := range obj.Items {
		track := item.(Track)
		tracks = append(tracks, &track)
	}

	return tracks, nil
}

func (s *service) GetAlbums(token string, ids []string) ([]*Album, error) {
	idsParam := strings.Join(ids, ",")
	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/albums?ids=%s", idsParam)

	res, err := s.makeRequest(token, http.MethodGet, spotifyURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "[GetAlbums]: unable to make request")
	}

	var obj Paging
	err = json.Unmarshal(res, &obj)
	if err != nil {
		return nil, errors.Wrap(err, "[GetAlbums]: unable to unmarshal response body")
	}

	albums := []*Album{}
	for _, item := range obj.Items {
		album := item.(Album)
		albums = append(albums, &album)
	}

	return albums, nil
}