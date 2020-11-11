package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	defaultTimeout = 30
	scope          = "user-read-recently-played playlist-modify-public playlist-read-collaborative user-read-recently-played user-top-read user-library-read"
	redirectURL    = "https://bot-bk.website/spotifyCallback"

	AuthState = "spotify-auth-state"
)

var (
	errorUnableToGetToken = errors.New("unable to get token from spotify")
)

type SpotifyService interface {
	Login(state string) error
}

type spotifyService struct {
	ClientID    string
	ClintSecret string
	repository  Repository
}

type requestTokenBody struct {
	Code        string `json:"code"`
	GrantType   string `json:"grant_type"`
	RedirectURI string `json:"redirect_uri"`
}

type responseTokenBody struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
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

	return encoded
}

func (s *spotifyService) Login(state string) error {
	url := "https://accounts.spotify.com/authorize"

	client := &http.Client{
		Timeout: time.Second * defaultTimeout,
	}

	req, err := http.NewRequest("GET", url, nil)

	q := req.URL.Query()
	q.Add("client_id", s.ClientID)
	q.Add("response_type", "code")
	q.Add("redirect_uri", redirectURL)
	q.Add("state", state)
	q.Add("scope", scope)
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Err(errors.Wrap(err, "[Login]: unable to close response body"))
		}
	}()
	if err != nil {
		return errors.Wrap(err, "[Login]: unable to get authorized from spotify")
	}

	return nil
}

func (s *spotifyService) GetTokenRequest(uid, code string) error {
	url := "https://accounts.spotify.com/api/token"
	body := requestTokenBody{
		Code:        code,
		GrantType:   "authorization_code",
		RedirectURI: redirectURL,
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return errors.Wrap(err, "[GetTokenRequest]: unable to marshal request body")
	}

	client := &http.Client{
		Timeout: time.Second * defaultTimeout,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return errors.Wrap(err, "[GetTokenRequest]: unable to create request")
	}
	req.Header.Add("Authorization", s.newAuthHeader())

	res, err := client.Do(req)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Err(errors.Wrap(err, "[GetTokenRequest]: unable to close response body"))
		}
	}()
	if err != nil {
		return errors.Wrap(err, "[GetTokenRequest]: unable to get token from spotify")
	}

	accToken, refToekn, err := s.GetTokenResponse(res)
	acc := Account{
		UID:          uid,
		AccessToken:  accToken,
		RefreshToken: refToekn,
	}
	if _, err = s.repository.CreateAccount(acc); err != nil {
		return errors.Wrap(err, "[GetTokenRequest]: unable to create account")
	}

	return nil
}

func (s *spotifyService) GetTokenResponse(res *http.Response) (string, string, error) {
	if res.StatusCode != http.StatusOK {
		return "", "", errorUnableToGetToken
	}

	var body responseTokenBody
	err := json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return "", "", errors.Wrap(err, "[GetTokenResponse]: unable to unmarshal response body")
	}

	accessToken := body.AccessToken
	refreshToken := body.RefreshToken

	return accessToken, refreshToken, nil
}
