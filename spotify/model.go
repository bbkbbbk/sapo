package spotify

type User struct {
	ID     string       `json:"id"`
	Name   string       `json:"display_name"`
	Email  string       `json:"email"`
	Images []Image      `json:"images"`
	URL    ExternalURLs `json:"external_urls"`
}

type Track struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Artists    []Artist     `json:"artists"`
	Album      Album        `json:"album"`
	URL        ExternalURLs `json:"external_urls"`
	SpotifyURI string       `json:"uri"`
}

type Artist struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Genres     []string     `json:"genres"`
	Images     []Image      `json:"images"`
	URL        ExternalURLs `json:"external_urls"`
	SpotifyURI string       `json:"uri"`
}

type Album struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Label      string            `json:"label"`
	Artists    []Artist          `json:"artists"`
	Genres     []string          `json:"genres"`
	Images     []Image           `json:"images"`
	Tracks     TrackPagingObject `json:"tracks"`
	URL        ExternalURLs      `json:"external_urls"`
	SpotifyURI string            `json:"uri"`
}

type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type ExternalURLs struct {
	URL string `json:"spotify"`
}

type TrackPagingObject struct {
	Items    []Track `json:"items"`
	Limit    int     `json:"limit"`
	Offset   int     `json:"offset"`
	Next     string  `json:"next"`
	Previous string  `json:"previous"`
	Total    int     `json:"total"`
}

// SimplifiedObject represents simplified version of track, artist, album, and playlist
type SimplifiedObject struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	URI          string       `json:"uri"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type SimplifiedTracks struct {
	Tracks []SimplifiedObject `json:"tracks"`
}

type Playlist struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Images       []Image      `json:"images"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

// PlayHistoryTrack represents items inside response from getting user currently played tracks
type PlayHistoryTrack struct {
	Track    SimplifiedObject `json:"track"`
	PlayedAt string           `json:"played_at"`
}

// PlayHistoryObject represent object response from getting user currently played tracks
type PlayHistoryObject struct {
	Items []PlayHistoryTrack `json:"items"`
}
