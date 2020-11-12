package server

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
	redirectURL    = "https://sapo-wb87j.ondigitalocean.app/spotify-callback"

	AuthState = "spotify-auth-state"
)

var (
	errorUnableToGetToken = errors.New("unable to get token from spotify")
)

type SpotifyService interface {
	GetAuthURL(state string) string
	RequestToken(code string) (string, string, error)
}

type spotifyService struct {
	ClientID    string
	ClintSecret string
}

type responseTokenBody struct {
	AccessToken    string `json:"access_token"`
	TokenType      string `json:"token_type"`
	Scope          string `json:"scope"`
	ExpirationTime int    `json:"expires_in"`
	RefreshToken   string `json:"refresh_token"`
}

func NewSpotifyService(id, secret string) SpotifyService {
	return &spotifyService{
		ClientID:    id,
		ClintSecret: secret,
	}
}

func (s *spotifyService) newAuthHeader() string {
	raw := fmt.Sprintf("%s:%s", s.ClientID, s.ClintSecret)
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))
	authHeader := fmt.Sprintf("Basic %s", encoded)

	return authHeader
}

func (s *spotifyService) GetAuthURL(state string) string {
	spotifyURL := "https://accounts.spotify.com/authorize"
	scope := url.QueryEscape(scopes)
	url := fmt.Sprintf("%s?client_id=%s&scope=%s&response_type=code&redirect_uri=%s&state=%s", spotifyURL, s.ClientID, scope, redirectURL, state)

	return url
}

func (s *spotifyService) RequestToken(code string) (string, string, error) {
	spotifyURL := "https://accounts.spotify.com/api/token"

	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", redirectURL)

	req, err := http.NewRequest("POST", spotifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", errors.Wrap(err, "[RequestToken]: unable to create request")
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
			logrus.Warn("[RequestToken]: unable to close response body", err)
		}
	}()
	if err != nil {
		return "", "", errors.Wrap(err, "[RequestToken]: unable to request token from spotify")
	}

	accToken, refToken, err := s.getTokenFromResponse(res)
	if err != nil {
		return "", "", errors.Wrap(err, "[RequestToken]: unable to get token response from spotify")
	}

	return accToken, refToken, nil
}

func (s *spotifyService) getTokenFromResponse(res *http.Response) (string, string, error) {
	if res.StatusCode != http.StatusOK {
		logrus.Warn("[getTokenFromResponse]: unable to make a success response")
		return "", "", errorUnableToGetToken
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "[getTokenFromResponse]: unable to read response body")
	}

	var token responseTokenBody
	err = json.Unmarshal(body, &token)
	if err != nil {
		return "", "", errors.Wrap(err, "[getTokenFromResponse]: unable to unmarshal response body")
	}

	accessToken := token.AccessToken
	refreshToken := token.RefreshToken

	return accessToken, refreshToken, nil
}
