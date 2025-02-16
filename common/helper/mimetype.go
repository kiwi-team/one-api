package helper

import (
	"net/http"

	"github.com/gabriel-vasile/mimetype"
)

func DetectFileFromURL(url string) (*mimetype.MIME, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return mimetype.DetectReader(resp.Body)
}
