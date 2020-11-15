package spotify

type User struct {
	ID           string       `json:"id"`
	Name         string       `json:"display_name"`
	Email        string       `json:"email"`
	Images       []Image      `json:"images"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type Track struct {
	ID           string             `json:"id"`
	Name         string             `json:"display_name"`
	Artists      []SimplifiedObject `json:"artists"`
	Album        SimplifiedObject   `json:"album"`
	Duration     int                `json:"duration_ms"`
	PreviewURL   string             `json:"preview_url"`
	ExternalURLs ExternalURLs       `json:"external_urls"`
}

type Artist struct {
	ID           string       `json:"id"`
	Name         string       `json:"display_name"`
	Images       []Image      `json:"images"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type Album struct {
	ID           string             `json:"id"`
	Name         string             `json:"display_name"`
	Label        string             `json:"label"`
	Artists      []SimplifiedObject `json:"artists"`
	Tracks       []SimplifiedObject `json:"tracks"`
	Images       []Image            `json:"images"`
	ExternalURLs ExternalURLs       `json:"external_urls"`
}

type Playlist struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Images       []Image      `json:"images"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type ExternalURLs struct {
	URL string `json:"spotify"`
}

// SimplifiedObject represents simplified version of track, artist, album, and playlist
type SimplifiedObject struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	URI          string       `json:"uri"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type Tracks struct {
	Items []SimplifiedObject `json:"tracks"`
}

type PlayingHistory struct {
	Track    SimplifiedObject `json:"track"`
	PlayedAt string           `json:"played_at"`
}

type Paging struct {
	Items []interface{} `json:"items"`
}
