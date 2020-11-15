package message

import (
	"fmt"
)

type BubbleWithButton struct {
	AltText     string
	ImageURL    string
	Header      string
	Text        string
	ButtonLabel string
	URLAction   string
	Color       string
}

func NewBubbleWithButton(altText, imgUrl, header, text, buttonLabel, urlAction, color string) Flex {
	return &BubbleWithButton{
		AltText: altText,
		ImageURL: imgUrl,
		Header: header,
		Text: text,
		ButtonLabel: buttonLabel,
		URLAction: urlAction,
		Color: color,
	}
}

func (f *BubbleWithButton) ToComponent() string {
	cover := fmt.Sprintf(`{
					"type": "image",
					"url": "%s",
					"size": "full",
					"aspectMode": "cover",
					"gravity": "center"
				  }`, f.ImageURL)

	header := fmt.Sprintf(`{
						"type": "box",
						"layout": "vertical",
						"contents": [
						  {
							"type": "text",
							"text": "%s",
							"color": "#ffffff",
							"weight": "bold",
							"size": "lg"
						  }
						]
					  }`, f.Header)

	text := fmt.Sprintf(`{
						"type": "box",
						"layout": "vertical",
						"contents": [
						  {
							"type": "text",
							"text": "%s",
							"color": "#969696",
							"size": "xs"
						  }
						]
					  }`, f.Text)

	button := fmt.Sprintf(`{
						"type": "button",
						"action": {
						  "type": "uri",
						  "label": "%s",
						  "uri": "%s"
						},
						"color": "#ffffff",
						"offsetBottom": "5px"
					  }`, f.ButtonLabel, f.URLAction)

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
				  }`, header, text, button, f.Color)

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
				}`, f.AltText, bubble)

	return flex
}
