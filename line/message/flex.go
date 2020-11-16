package message

import (
	"fmt"
	"strings"
)

type BubbleWithButton struct {
	AltText     string
	Header      string
	Text        string
	ButtonLabel string
	URLAction   string
	ImageURL    string
	Color       string
}

func NewBubbleWithButton(altText, header, text, buttonLabel, urlAction, imgUrl, color string) Flex {
	return &BubbleWithButton{
		AltText:     altText,
		Header:      header,
		Text:        text,
		ButtonLabel: buttonLabel,
		URLAction:   urlAction,
		ImageURL:    imgUrl,
		Color:       color,
	}
}

func (b *BubbleWithButton) ToComponent() string {
	cover := fmt.Sprintf(`{
					"type": "image",
					"url": "%s",
					"size": "full",
					"aspectMode": "cover",
					"gravity": "center"
				  }`, b.ImageURL)
	header := fmt.Sprintf(`{
						"type": "box",
						"layout": "vertical",
						"contents": [
						  {
							"type": "text",
							"text": "%s",
							"color": "#ffffff",
							"weight": "bold",
							"size": "sm"
						  }
						]
					  }`, b.Header)
	text := fmt.Sprintf(`{
						"type": "box",
						"layout": "vertical",
						"contents": [
						  {
							"type": "text",
							"text": "%s",
							"color": "#969696",
							"size": "xxs"
						  }
						]
					  }`, b.Text)
	button := fmt.Sprintf(`{
						"type": "button",
						"action": {
						  "type": "uri",
						  "label": "%s",
						  "uri": "%s"
						},
						"color": "#ffffff",
						"offsetBottom": "5px"
					  }`, b.ButtonLabel, b.URLAction)
	footer := fmt.Sprintf(`{
					"type": "box",
					"layout": "vertical",
					"contents": [%s,%s,%s
					],
					"height": "100px",
					"backgroundColor": "#%s",
					"position": "absolute",
					"offsetBottom": "0px",
					"offsetStart": "0px",
					"offsetEnd": "0px",
					"paddingAll": "10px"
				  }`, header, text, button, b.Color)
	bubble := fmt.Sprintf(
		`{
			  "type": "bubble",
			  "size": "kilo",
			  "body": {
				"type": "box",
				"layout": "vertical",
				"contents": [%s,%s],
				"paddingAll": "0px"
			  }
			}`, cover, footer)
	flex := fmt.Sprintf(
		`{
				  "type": "flex",
				  "altText": "%s",
				  "contents": %s
				}`, b.AltText, bubble)

	return flex
}

func (b *BubbleWithButton) ToJson() []byte {
	return []byte(b.ToComponent())
}

type BubbleReceipt struct {
	AltText string
	TopText string
	Header  string
	Text    string
	Items   []BoxWithImage
}

type BoxWithImage struct {
	Header   string
	Text     string
	LeftText string
	ImageURL string
	URL      string
}

func NewBubbleReceipt(altText, topText, header, text string, items []BoxWithImage) Flex {
	return &BubbleReceipt{
		AltText: altText,
		TopText: topText,
		Header:  header,
		Text:    text,
		Items:   items,
	}
}

func (b *BubbleReceipt) ToComponent() string {
	boxes := []string{}
	for _, item := range b.Items {
		boxes = append(boxes, item.ToComponent())
	}

	topText := fmt.Sprintf(`{
				  "type": "text",
				  "text": "%s",
				  "weight": "bold",
				  "color": "#2FA6E9",
				  "size": "sm"
				}`, b.TopText)
	header := fmt.Sprintf(`{
				  "type": "text",
				  "text": "%s",
				  "weight": "bold",
				  "size": "xxl",
				  "margin": "md",
				  "color": "#373C41"
				}`, b.Header)
	text := fmt.Sprintf(`{
				  "type": "text",
				  "text": "%s",
				  "size": "xs",
				  "color": "#969696",
				  "wrap": true,
  				  "offsetTop": "5px"
				}`, b.Text)
	bubble := fmt.Sprintf(`{
				"type": "bubble",
				"body": {
				  "type": "box",
				  "layout": "vertical",
				  "contents": [%s,%s,%s,
					{
					  "type": "separator",
					  "margin": "xxl"
					},
					{
					  "type": "box",
					  "layout": "vertical",
					  "margin": "xxl",
					  "spacing": "sm",
					  "contents": [%s]
					}
				  ]
				}
			  }`, topText, header, text, strings.Join(boxes, ","))
	flex := fmt.Sprintf(
		`{
				  "type": "flex",
				  "altText": "%s",
				  "contents": %s
				}`, b.AltText, bubble)

	return flex
}

func (b *BoxWithImage) ToComponent() string {
	image := fmt.Sprintf(`{
				  "type": "image",
				  "url": "%s",
				  "size": "50px",
				  "align": "start",
				  "aspectRatio": "1:1"
				}`, b.ImageURL)
	header := fmt.Sprintf(`{
				  "type": "text",
				  "text": "%s",
				  "color": "#373C41",
				  "size": "sm",
				  "weight": "bold",
				  "align": "start"
				}`, b.Header)
	leftText := fmt.Sprintf(`{
				  "type": "text",
				  "text": "%s",
				  "color": "#969696",
				  "size": "xxs",
				  "align": "end"
				}`, b.LeftText)
	text := fmt.Sprintf(`{
				  "type": "text",
				  "text": "%s",
				  "color": "#969696",
				  "size": "xxs"
				}`, b.Text)
	action := fmt.Sprintf(`{
                "type": "uri",
                "label": "action",
                "uri": "%s"
              }`, b.URL)
	box := fmt.Sprintf(`{
			  "type": "box",
			  "layout": "horizontal",
			  "contents": [
				%s,
				{
				  "type": "box",
				  "layout": "vertical",
				  "contents": [
					{
					  "type": "box",
					  "layout": "baseline",
					  "contents": [%s,%s],
					  "width": "200px"
					},
					%s
				  ],
				  "position": "absolute",
				  "offsetStart": "60px",
				  "offsetTop": "5px"
				}
			  ],
			  "action": %s,
			  "paddingBottom": "10px"
			}`, image, header, leftText, text, action)

	return box
}

func (b *BubbleReceipt) ToJson() []byte {
	return []byte(b.ToComponent())
}
