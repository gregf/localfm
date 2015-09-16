package commands

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (env *Env) Daemon(cmd *cobra.Command, args []string) {
	ticker := time.NewTicker(1 * time.Minute)
	fmt.Printf("LocalFM Deamon %s Started\n", localFMVersion)
	for _ = range ticker.C {
		user := viper.GetString("main.lastfm_username")
		apiKey := viper.GetString("main.lastfm_apikey")
		epoch, err := env.db.FindLastListen()
		if err != nil {
			log.Fatal("Error parsing time:", err)
		}

		lastPage, err := TotalPages(baseURL, user, apiKey, limit, epoch)
		if err != nil {
			log.Fatal("Could not obtain TotalPages:", err)
		}
		firstPage := 1

		for i := lastPage; i >= firstPage; i-- {
			url := fmt.Sprintf("%s&api_key=%s&user=%s&page=%d&limit=%d&from=%d", baseURL, apiKey, user, i, limit, epoch)

			resp, err := FetchBody(url)
			if err != nil {
				log.Fatal(err)
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
				if t.NowPlaying {
					continue
				}
				dt, err := time.Parse("02 Jan 2006, 15:04", t.Date)
				if err != nil {
					log.Printf("Error parsing time on %s / %s - %s / %s: %s\n", t.Artist, t.Album, t.Name, t.Date, err)
					continue
				}
				if dt.IsZero() {
					log.Println("Time is Zero")
					continue
				}
				env.db.AddArtist(t.Artist)
				env.db.AddTrack(t.Artist, t.Album, t.Name, dt)
			}
		}
	}
}
