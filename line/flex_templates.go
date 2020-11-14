package line

import (
	"fmt"
)

type FlexTemplate struct {
	Header      string
	Text        string
	ButtonLabel string
	URLAction   string
	ImageURL    string
	Color       string
}

func (f *FlexTemplate) ToJson() []byte {
	template := fmt.Sprintf(
		`{
			  "type": "bubble",
			  "size": "kilo",
			  "body": {
				"type": "box",
				"layout": "vertical",
				"contents": [
				  {
					"type": "image",
					"url": "%s",
					"size": "full",
					"aspectMode": "cover",
					"gravity": "center"
				  },
				  {
					"type": "box",
					"layout": "vertical",
					"contents": [
					  {
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
					  },
					  {
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
					  },
					  {
						"type": "button",
						"action": {
						  "type": "uri",
						  "label": "%s",
						  "uri": "%s"
						},
						"color": "#ffffff",
						"offsetBottom": "5px"
					  }
					],
					"height": "100px",
					"backgroundColor": "#%s",
					"position": "absolute",
					"offsetBottom": "0px",
					"offsetStart": "0px",
					"offsetEnd": "0px",
					"paddingAll": "10px"
				  }
				],
				"paddingAll": "0px"
			  }
			}`, f.ImageURL, f.Header, f.Text, f.ButtonLabel, f.URLAction, f.Color)

	return []byte(template)
}
