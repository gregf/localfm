package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/gohome"
	"github.com/jinzhu/gorm"
	// required by gorm
	_ "github.com/mattn/go-sqlite3"
)

// Datastore interface
type Datastore interface {
	AddArtist(name string)
	AddTrack(artist, album, title string, date time.Time)
	FindLastListen() (int64, error)
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
	if db.NewRecord(&artist) {
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

	if db.NewRecord(&track) {
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
