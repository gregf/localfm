package commands

import "encoding/xml"

var (
	baseURL = "http://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks"
	limit   = 150
)

type LFM struct {
	XMLName      xml.Name     `xml:"lfm"`
	Status       string       `xml:"status,attr"`
	RecentTracks RecentTracks `xml:"recenttracks"`
}
type RecentTracks struct {
	XMLName    xml.Name `xml:"recenttracks"`
	User       string   `xml:"user,attr"`
	Page       int      `xml:"page,attr"`
	PerPage    int      `xml:"perPage,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Total      int      `xml:"total,attr"`
	Tracks     []Track  `xml:"track"`
}

type Track struct {
	XMLName    xml.Name `xml:"track"`
	Artist     string   `xml:"artist"`
	Album      string   `xml:"album"`
	Name       string   `xml:"name"`
	Date       string   `xml:"date"`
	NowPlaying bool     `xml:nowplaying`
}
