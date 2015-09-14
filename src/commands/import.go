package commands

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (env *Env) Import(cmd *cobra.Command, args []string) {
	user := viper.GetString("main.lastfm_username")
	apiKey := viper.GetString("main.lastfm_apikey")

	lastPage, err := TotalPages(baseURL, user, apiKey, limit, 0)
	if err != nil {
		log.Fatal("Could not obtain TotalPages:", err)
	}
	firstPage := 1
	totalScrobbles, err := TotalScrobbles(baseURL, user, apiKey, limit, 0)
	if err != nil {
		log.Fatal("Could not obtain Total Scrobbles:", err)
	}

	n := 1
	for i := lastPage; i >= firstPage; i-- {
		url := fmt.Sprintf("%s&api_key=%s&user=%s&page=%d&limit=%d", baseURL, apiKey, user, i, limit)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var l LFM
		err = xml.Unmarshal(body, &l)
		if err != nil {
			log.Fatal(err)
		}

		totalItems := (len(l.RecentTracks.Tracks) - 1)
		for i := totalItems; i >= 0; i-- {
			t := l.RecentTracks.Tracks[i]
			dt, err := time.Parse("02 Jan 2006, 15:04", t.Date)
			if err != nil {
				log.Println("Error parsing time", err)
			}
			env.db.AddArtist(t.Artist)
			if env.db.AddTrack(t.Artist, t.Album, t.Name, dt) {
				fmt.Printf("\033[H\033[2J%d/%d %s / %s - %s", n, totalScrobbles, t.Artist, t.Album, t.Name)
				n++
			}
		}
	}
}
