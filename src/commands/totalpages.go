package commands

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func TotalPages(baseURL, user, apiKey string, limit int, from int64) (int, error) {
	var url string
	if from == 0 {
		url = fmt.Sprintf("%s&api_key=%s&user=%s&page=1&limit=%d", baseURL, apiKey, user, limit)
	} else {
		url = fmt.Sprintf("%s&api_key=%s&user=%s&page=1&limit=%d&from=%d", baseURL, apiKey, user, limit, from)
	}
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return 0, err
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

	lastPage := l.RecentTracks.TotalPages
	return lastPage, nil
}
