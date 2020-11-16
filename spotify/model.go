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
	Name         string             `json:"name"`
	Artists      []SimplifiedObject `json:"artists"`
	Album        SimplifiedObject   `json:"album"`
	Duration     int                `json:"duration_ms"`
	PreviewURL   string             `json:"preview_url"`
	ExternalURLs ExternalURLs       `json:"external_urls"`
	URI          string             `json:"uri"`
}

type ExternalURLs struct {
	URL string `json:"spotify"`
}

type Artist struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Images       []Image      `json:"images"`
	ExternalURLs ExternalURLs `json:"external_urls"`
	URI          string       `json:"uri"`
}

type Album struct {
	ID           string             `json:"id"`
	Name         string             `json:"display_name"`
	Label        string             `json:"label"`
	Artists      []SimplifiedObject `json:"artists"`
	Tracks       []Track            `json:"tracks.items"`
	Images       []Image            `json:"images"`
	ExternalURLs ExternalURLs       `json:"external_urls"`
	URI          string             `json:"uri"`
}

type Playlist struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Images       []Image      `json:"images"`
	ExternalURLs ExternalURLs `json:"external_urls"`
	URI          string       `json:"uri"`
}

type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// SimplifiedObject represents simplified version of track, artist, album, and playlist
type SimplifiedObject struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	URI          string       `json:"uri"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type Tracks struct {
	Items []Track `json:"tracks"`
}

type Albums struct {
	Items []Album `json:"albums"`
}

type PlayingHistory struct {
	Track    SimplifiedObject `json:"track"`
	PlayedAt string           `json:"played_at"`
}

type PlayingHistoryItems struct {
	PlayingHistories []PlayingHistory `json:"items"`
}

type TrackItems struct {
	Tracks []Track `json:"items"`
}

type ArtistItems struct {
	Artists []Artist `json:"items"`
}

type AlbumItems struct {
	Albums []Album `json:"items"`
}
