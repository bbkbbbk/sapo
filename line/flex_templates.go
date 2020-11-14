package line

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type FlexTemplate struct {
	Header string
	Text string
	ButtonLabel string
	URLAction string
	ImageURL string
	Color string
}

func (f *FlexTemplate) ToJson() ([]byte, error) {
	template := fmt.Sprintf(
		`{
				  "type": "bubble",
				  "body": {
					"type": "box",
					"layout": "vertical",
					"contents": [
					  {
						"type": "image",
						"url": "%s",
						"size": "full",
						"aspectMode": "cover",
						"aspectRatio": "3:4",
						"gravity": "center"
					  },
					  {
						"type": "box",
						"layout": "vertical",
						"contents": [],
						"position": "absolute",
						"background": {
						  "type": "linearGradient",
						  "angle": "0deg",
						  "endColor": "#00000000",
						  "startColor": "#00000099"
						},
						"width": "100%%",
						"height": "40%%",
						"offsetBottom": "0px",
						"offsetStart": "0px",
						"offsetEnd": "0px"
					  },
					  {
						"type": "box",
						"layout": "horizontal",
						"contents": [
						  {
							"type": "box",
							"layout": "vertical",
							"contents": [
							  {
								"type": "box",
								"layout": "horizontal",
								"contents": [
								  {
									"type": "text",
									"text": "%s",
									"size": "lg",
									"color": "#ffffff"
								  }
								]
							  },
							  {
								"type": "box",
								"layout": "baseline",
								"contents": [
								  {
									"type": "text",
									"text": "%s",
									"color": "#a9a9a9",
									"size": "xs"
								  }
								],
								"spacing": "xs"
							  },
							  {
								"type": "box",
								"layout": "horizontal",
								"contents": [
								  {
									"type": "button",
									"action": {
									  "type": "uri",
									  "label": "%s",
									  "uri": "%s"
									},
									"color": "#ffffff"
								  }
								]
							  }
							],
							"spacing": "xs"
						  }
						],
						"position": "absolute",
						"offsetBottom": "0px",
						"offsetStart": "0px",
						"offsetEnd": "0px",
						"paddingAll": "10px",
						"backgroundColor": "%s"
					  }
					],
					"paddingAll": "0px"
				  }
				}`, f.ImageURL, f.Header, f.Text, f.ButtonLabel, f.URLAction, f.Color)

	jsonTemplate, err := json.Marshal(template)
	if err != nil {
		return nil, errors.Wrap(err, "[ToFlexComponent]: unable to marshal template")
	}

	return jsonTemplate, nil
}
