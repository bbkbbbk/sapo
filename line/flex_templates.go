package line

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

func CreatePlaylistFlexTemplate(name, des, url, imgUrl string) ([]byte, error) {
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
									  "label": "go to playlist",
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
						"backgroundColor": "#373C41CC"
					  }
					],
					"paddingAll": "0px"
				  }
				}`, imgUrl, name, des, url)
	b, err := json.Marshal(template)
	if err != nil {
		return nil, errors.Wrap(err, "[CreatePlaylistFlex]: unable to marshal template")
	}

	return b, nil
}
