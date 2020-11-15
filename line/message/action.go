package message

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type Flex interface {
	ToComponent() string
	ToJson() []byte
}

type Reply struct {
	ReplyToken string
	Message    Flex
}

func (r *Reply) ToJson() []byte {
	msg := fmt.Sprintf(`{
		"replyToken":"%s",
		"messages":[%s]
	}`, r.ReplyToken, r.Message.ToComponent())

	logrus.Info(msg)

	return []byte(msg)
}

type Push struct {
	ToID    string
	Message Flex
}

func (r *Push) ToJson() []byte {
	msg := fmt.Sprintf(`{
		"to":"%s",
		"messages":[%s]
	}`, r.ToID, r.Message.ToComponent())

	logrus.Info(msg)

	return []byte(msg)
}
