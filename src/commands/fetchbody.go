package commands

import (
	"fmt"
	"net/http"
)

// FetchBody fetches a http response body for a specific url
func FetchBody(url string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	useragent := fmt.Sprintf("LocalFM %s", localFMVersion)
	req.Header.Add("User-Agent", useragent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
