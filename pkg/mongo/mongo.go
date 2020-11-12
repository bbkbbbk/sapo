package mongo

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	errorUnableToConnectToDatabase = errors.New("unable to connect database")
)

type Config struct {
	AuthSource string
	Database   string
	Host       string
	Username   string
	Password   string
}

func NewMongo(c Config) *mongo.Database {
	logrus.Print("[main]: initializing mongo connection")

	uri := fmt.Sprintf("mongodb://%s:%s@%s:27017/?authSource=%v", c.Username, c.Password, c.Host, c.AuthSource)

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		logrus.Warn(errorUnableToConnectToDatabase)
	}
	db := client.Database(c.Database)

	return db
}
