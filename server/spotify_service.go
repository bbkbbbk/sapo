package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	scope          = "user-read-recently-played playlist-modify-public playlist-read-collaborative user-read-recently-played user-top-read user-library-read"
	redirectURL    = "http://localhost:8080/spotify-callback"
	AuthState      = "spotify-auth-state"
)

var (
	errorUnableToGetToken = errors.New("unable to get token from spotify")
)

type SpotifyService interface {
	GetAuthURL(state string) string
	GetTokenRequest(uid, code string) error
	GetTokenResponse(res *http.Response) (string, string, error)
}

type spotifyService struct {
	ClientID    string
	ClintSecret string
	repository  Repository
}

type responseTokenBody struct {
	AccessToken    string `json:"access_token"`
	TokenType      string `json:"token_type"`
	Scope          string `json:"scope"`
	ExpirationTime int    `json:"expires_in"`
	RefreshToken   string `json:"refresh_token"`
}

func NewSpotifyService(id, secret string, r Repository) SpotifyService {
	return &spotifyService{
		ClientID:    id,
		ClintSecret: secret,
		repository:  r,
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
	scopes := url.QueryEscape(scope)
	url := fmt.Sprintf("%s?client_id=%s&scope=%s&response_type=code&redirect_uri=%s&state=%s", spotifyURL, s.ClientID, scopes, redirectURL, state)

	return url
}

func (s *spotifyService) GetTokenRequest(uid, code string) error {
	spotifyURL := "https://accounts.spotify.com/api/token"

	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", redirectURL)

	req, err := http.NewRequest("POST", spotifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return errors.Wrap(err, "[GetTokenRequest]: unable to create request")
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
			logrus.Warnf("[GetTokenRequest]: unable to close response body", err)
		}
	}()
	if err != nil {
		return errors.Wrap(err, "[GetTokenRequest]: unable to get token from spotify")
	}

	accToken, refToken, err := s.GetTokenResponse(res)
	acc := Account{
		UID:          uid,
		AccessToken:  accToken,
		RefreshToken: refToken,
	}
	if _, err := s.repository.CreateAccount(acc); err != nil {
		return errors.Wrap(err, "[GetTokenRequest]: unable to create account")
	}

	return nil
}

func (s *spotifyService) GetTokenResponse(res *http.Response) (string, string, error) {
	if res.StatusCode != http.StatusOK {
		logrus.Warn("[GetTokenResponse]: unable to make a success response")
		return "", "", errorUnableToGetToken
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}

	var token responseTokenBody
	err = json.Unmarshal(body, &token)
	if err != nil {
		return "", "", errors.Wrap(err, "[GetTokenResponse]: unable to unmarshal response body")
	}

	accessToken := token.AccessToken
	refreshToken := token.RefreshToken

	return accessToken, refreshToken, nil
}
