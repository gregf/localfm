package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/caarlos0/gohome"
	"github.com/dustin/go-humanize"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	// required by gorm
	_ "github.com/mattn/go-sqlite3"
)

// Datastore interface
type Datastore interface {
	AddArtist(name string) bool
	AddTrack(artist, album, title string, date time.Time) bool
	FindLastListen() (int64, error)
	RecentTracks() (string, error)
	Scrobbles() string
	TopArtists() (string, error)
	TopAlbums() (string, error)
	TopSongs() (string, error)
}

// DB struct
type DB struct {
	gorm.DB
}

const appName = "localfm"

// Artist struct
type Artist struct {
	ID     int    `sql:"index"`
	Name   string `sql:"unique_index"`
	Tracks []Track
}

// Track struct
type Track struct {
	ID       int `sql:"index"`
	ArtistID int
	Title    string
	Artist   string
	Album    string
	Date     time.Time `sql:"unique_index"`
}

func databasePath() (path string) {
	path = gohome.Cache(appName)
	os.MkdirAll(path, 0755)
	return filepath.Join(path, "cache.db")
}

// NewDB establishes a connection with the database and sets the DB struct
func NewDB() (*DB, error) {
	var err error
	db, err := gorm.Open("sqlite3", databasePath())
	if err != nil {
		return nil, err
	}

	db.LogMode(false)
	db.CreateTable(&Artist{})
	db.CreateTable(&Track{})
	db.AutoMigrate(&Artist{}, &Track{})

	return &DB{db}, nil
}

// AddArtist Inserts a new artist into the database
func (db *DB) AddArtist(name string) bool {
	artist := Artist{
		Name: name,
	}
	if db.NewRec("artists", "name", name) {
		db.Create(&artist)
		return true
	}
	return false
}

// findArtistID by artist name
func (db *DB) findArtistID(artist string) (artistID int) {
	row := db.Table("artists").
		Where("name = ?", artist).
		Select("id").
		Row()
	row.Scan(&artistID)
	return artistID
}

// AddTrack inserts a track into the database
func (db *DB) AddTrack(artist, album, title string, date time.Time) bool {
	artistID := db.findArtistID(artist)

	track := Track{
		Title:    title,
		Artist:   artist,
		Album:    album,
		ArtistID: artistID,
		Date:     date,
	}

	if db.NewRec("tracks", "date", date.String()) {
		db.Create(&track)
		return true
	}
	return false
}

func (db *DB) FindLastListen() (int64, error) {
	var date time.Time

	row := db.Table("tracks").
		Order("id desc").
		Limit(1).
		Select("date").
		Row()
	row.Scan(&date)

	t, err := time.Parse("2006-01-02 15:04:05 -0700 UTC", date.String())
	if err != nil {
		return 0, err
	}

	return t.UTC().Unix(), nil
}

// NewRec returns a bool depending on whether or not it could find a record
func (db *DB) NewRec(table, field, data string) bool {
	var d string
	if isADate(data) {
		t, err := time.Parse("2006-01-02 15:04:05 -0700 UTC", data)
		if err != nil {
			log.Fatalf("Problem parsing date: %s", err)
		}
		d = t.Format("2006-01-02 15:04:05")
	}

	var count int
	f := fmt.Sprintf("%s = ?", field)

	var row *gorm.DB
	if isADate(data) {
		row = db.Table(table).Where(f, d).Limit(1).Select(field).Count(&count)
	} else {
		row = db.Table(table).Where(f, data).Limit(1).Select(field).Count(&count)
	}

	row.Scan(&count)

	if count == 0 {
		return true
	}

	return false
}

//RecentTracks returns a string of recently played tracks.
func (db *DB) RecentTracks() (s string, err error) {
	var (
		title  string
		artist string
		date   time.Time
	)

	rows, err := db.Table("tracks").
		Select("title, artist, date").
		Order("id desc").
		Limit(viper.GetInt("main.recent_tracks")).
		Rows()
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var str []string
	for rows.Next() {
		rows.Scan(&title, &artist, &date)
		t, err := time.Parse("2006-01-02 15:04:05 -0700 UTC", date.String())
		if err != nil {
			return "", err
		}
		d := humanize.Time(t)
		str = append(str, fmt.Sprintf("%s - %s %s", title, artist, d))
	}
	return strings.Join(str, "\n"), nil
}

//Scrobbles returns a string with your userrname, number of scrobbles, artists,
// and your first play date.
func (db *DB) Scrobbles() (s string) {
	var (
		scrobblesCount int64
		artistsCount   int64
		date           time.Time
	)

	scrobbles := db.Table("tracks").Count(&scrobblesCount).Row()
	scrobbles.Scan(&scrobblesCount)

	artists := db.Table("artists").Count(&artistsCount)
	artists.Scan(&artistsCount)

	since := db.Table("tracks").Select("date").Order("date asc").Limit(1).Row()
	since.Scan(&date)

	d := date.Format("02 Jan 2006")

	s = fmt.Sprintf("%s     Scrobbles: %s     Artists: %s     Since: %s",
		viper.GetString("main.lastfm_username"),
		humanize.Comma(scrobblesCount),
		humanize.Comma(artistsCount),
		d)
	return s

}

// TopArtists returns a string of your top played artists
func (db *DB) TopArtists() (s string, err error) {
	type Result struct {
		Artist string
		Plays  int
	}

	sql := fmt.Sprintf("SELECT artist, COUNT(artist) AS plays FROM tracks GROUP BY artist ORDER BY COUNT(artist) DESC LIMIT %d;", viper.GetInt("main.top_artists"))
	rows, err := db.Raw(sql).Rows()
	if err != nil {
		return "", err
	}

	defer rows.Close()

	plays := make([]*Result, 0)
	for rows.Next() {
		play := new(Result)
		err := rows.Scan(&play.Artist, &play.Plays)
		if err != nil {
			return "", err
		}
		plays = append(plays, play)
	}

	var str []string
	for _, p := range plays {
		str = append(str, fmt.Sprintf("%s (%d plays)", p.Artist, p.Plays))
	}

	return strings.Join(str, "\n"), nil
}

// TopAlbums returns a string of your top played albums.
func (db *DB) TopAlbums() (s string, err error) {
	type Result struct {
		Artist string
		Album  string
		Plays  int
	}

	sql := fmt.Sprintf("SELECT artist, album, COUNT(album) AS plays FROM tracks GROUP BY album, artist ORDER BY COUNT(album) DESC LIMIT %d;", viper.GetInt("main.top_albums"))
	rows, err := db.Raw(sql).Rows()
	if err != nil {
		return "", err
	}

	defer rows.Close()

	plays := make([]*Result, 0)
	for rows.Next() {
		play := new(Result)
		err := rows.Scan(&play.Artist, &play.Album, &play.Plays)
		if err != nil {
			return "", err
		}
		plays = append(plays, play)
	}

	var str []string
	for _, p := range plays {
		str = append(str, fmt.Sprintf("%s - %s (%d plays)", p.Artist, p.Album, p.Plays))
	}

	return strings.Join(str, "\n"), nil
}

// TopSongs returns a string of your top played songs.
func (db *DB) TopSongs() (s string, err error) {
	type Result struct {
		Artist string
		Title  string
		Plays  int
	}

	sql := fmt.Sprintf("SELECT artist, title, COUNT(title) AS plays FROM tracks GROUP BY artist, title ORDER BY COUNT(title) DESC LIMIT %d;", viper.GetInt("main.top_songs"))
	rows, err := db.Raw(sql).Rows()
	if err != nil {
		return "", err
	}

	defer rows.Close()

	plays := make([]*Result, 0)
	for rows.Next() {
		play := new(Result)
		err := rows.Scan(&play.Artist, &play.Title, &play.Plays)
		if err != nil {
			return "", err
		}
		plays = append(plays, play)
	}

	var str []string
	for _, p := range plays {
		str = append(str, fmt.Sprintf("%s - %s (%d plays)", p.Artist, p.Title, p.Plays))
	}

	return strings.Join(str, "\n"), nil
}

// isADate returns a bool depending on whether a string is a date or just a string.
func isADate(date string) bool {
	_, err := time.Parse("2006-01-02 15:04:05 -0700 UTC", date)
	if err != nil {
		return false
	}

	return true
}
