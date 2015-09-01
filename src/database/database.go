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
	// required by gorm
	_ "github.com/mattn/go-sqlite3"
)

// Datastore interface
type Datastore interface {
	AddArtist(name string)
	AddTrack(artist, album, title string, date time.Time)
	FindLastListen() (int64, error)
	RecentTracks() string
	Scrobbles() string
	TopArtists() string
	TopAlbums() string
}

// DB struct
type DB struct {
	gorm.DB
}

const appName = "localfm"

// Podcast struct
type Artist struct {
	ID     int    `sql:"index"`
	Name   string `sql:"unique_index"`
	Tracks []Track
}

// Episode struct
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
func (db *DB) AddArtist(name string) {
	artist := Artist{
		Name: name,
	}
	if db.NewRec("artists", "name", name) {
		db.Create(&artist)
		fmt.Printf("Added New Artist: %s\n", name)
	}
}

// findPodcastID locates podcast ID by rssURL
func (db *DB) findArtistID(artist string) (artistID int) {
	row := db.Table("artists").
		Where("name = ?", artist).
		Select("id").
		Row()
	row.Scan(&artistID)
	return artistID
}

// AddTrack inserts a track into the database
func (db *DB) AddTrack(artist, album, title string, date time.Time) {
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
		fmt.Printf("Added New Track: %s / %s - %s\n", artist, album, title)
	}
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

func (db *DB) RecentTracks() (s string) {
	var (
		title  string
		artist string
		date   time.Time
	)

	rows, err := db.Table("tracks").
		Select("title, artist, date").
		Order("id desc").
		Limit(5).
		Rows()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var str []string
	for rows.Next() {
		rows.Scan(&title, &artist, &date)
		t, err := time.Parse("2006-01-02 15:04:05 -0700 UTC", date.String())
		if err != nil {
			log.Fatalf("Could not parse date: %s\n", err)
		}
		d := humanize.Time(t)
		str = append(str, fmt.Sprintf("%s - %s %s", title, artist, d))
	}
	return strings.Join(str, "\n")
}

func (db *DB) Scrobbles() (s string) {
	var (
		scrobblesCount int64
		artistsCount   int64
	)

	scrobbles := db.Table("tracks").Count(&scrobblesCount).Row()
	scrobbles.Scan(&scrobblesCount)

	artists := db.Table("artists").Count(&artistsCount)
	artists.Scan(&artistsCount)

	s = fmt.Sprintf("Scrobbles: %s\tArtists: %s", humanize.Comma(scrobblesCount), humanize.Comma(artistsCount))
	return s

}

func (db *DB) TopArtists() (s string) {
	type Result struct {
		Artist string
		Plays  int
	}

	rows, err := db.Raw("SELECT artist, COUNT(artist) AS plays FROM tracks GROUP BY artist ORDER BY COUNT(artist) DESC LIMIT 5;").Rows()
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	plays := make([]*Result, 0)
	for rows.Next() {
		play := new(Result)
		err := rows.Scan(&play.Artist, &play.Plays)
		if err != nil {
			log.Fatal(err)
		}
		plays = append(plays, play)
	}

	var str []string
	for _, p := range plays {
		str = append(str, fmt.Sprintf("%s (%d plays)", p.Artist, p.Plays))
	}

	return strings.Join(str, "\n")
}

func (db *DB) TopAlbums() (s string) {
	type Result struct {
		Artist string
		Album  string
		Plays  int
	}

	rows, err := db.Raw("SELECT artist, album, COUNT(album) AS plays FROM tracks GROUP BY album, artist ORDER BY COUNT(album) DESC LIMIT 5;").Rows()
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	plays := make([]*Result, 0)
	for rows.Next() {
		play := new(Result)
		err := rows.Scan(&play.Artist, &play.Album, &play.Plays)
		if err != nil {
			log.Fatal(err)
		}
		plays = append(plays, play)
	}

	var str []string
	for _, p := range plays {
		str = append(str, fmt.Sprintf("%s - %s (%d plays)", p.Artist, p.Album, p.Plays))
	}

	return strings.Join(str, "\n")
}

func isADate(date string) bool {
	_, err := time.Parse("2006-01-02 15:04:05 -0700 UTC", date)
	if err != nil {
		return false
	}

	return true
}
