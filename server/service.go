package server

import (
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/pkg/errors"
	"math/rand"
	"strings"
	"time"
)

type Service interface {
	RandomString(n int) string
	GetSpotifyAuthURL(state string) string
	CreateAccount(uid, code string) error
}

type service struct {
	botClient     *linebot.Client
	spotifyClient SpotifyService
	repository  Repository
}

func NewService(line *linebot.Client, spotify SpotifyService, repo Repository) Service {
	return &service{
		botClient: line,
		spotifyClient: spotify,
		repository: repo,
	}
}

func (s *service) RandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
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
	return s.spotifyClient.GetAuthURL(state)
}

func (s *service) CreateAccount(uid, code string) error {
	accToken, refToken, err := s.spotifyClient.RequestToken(code)
	if err != nil {
		return errors.Wrap(err, "[InsertUser]: unable to get token from spotify")
	}

	acc := Account{
		UID:          uid,
		AccessToken:  accToken,
		RefreshToken: refToken,
	}

	if _, err := s.repository.CreateAccount(acc); err != nil {
		return errors.Wrap(err, "[RequestToken]: unable to create account")
	}

	return nil
}
