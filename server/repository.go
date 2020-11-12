package server

import (
	"context"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	collNameAccounts = "accounts"
)

type Repository interface {
	CreateAccount(acc Account) (*Account, error)
}

type repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		db: db,
	}
}

type Account struct {
	UID          string     `json:"uid" bson:"uid"`
	AccessToken  string     `json:"accessToken" bson:"accessToken"`
	RefreshToken string     `json:"refreshToken" bson:"refreshToken"`
	CreatedAt    *time.Time `json:"createdAt" bson:"createdAt"`
}

func (r *repository) defaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*defaultTimeout)
}

func (r *repository) CreateAccount(acc Account) (*Account, error) {
	ctx, cancel := r.defaultContext()
	defer cancel()

	doc, err := bson.Marshal(acc)
	if err != nil {
		return nil, errors.Wrap(err, "[r.CreateAccount]: unable to marshal account")
	}

	_, err = r.db.Collection(collNameAccounts).InsertOne(ctx, doc)
	if err != nil {
		return nil, errors.Wrap(err, "[r.CreateAccount]: failed to insert account")
	}

	logrus.Printf("account created: %v", acc)

	return &acc, nil
}
