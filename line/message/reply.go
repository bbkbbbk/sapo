package message

import "fmt"

type Flex interface {
	ToComponent() string
}

type Reply struct {
	ReplyToken string
	Message Flex
}

func (r *Reply) ToJson() []byte {
	msg := fmt.Sprintf(`{
		"replyToken":"%s",
		"messages":[%s]
	}`, r.ReplyToken, r.Message.ToComponent())

	return []byte(msg)
}