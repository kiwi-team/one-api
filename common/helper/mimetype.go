package helper

import (
	"encoding/base64"
	"io"
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

func DetectFileAndBase64File(url string) (*mimetype.MIME, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// 读取 resp.Body 的内容到缓冲区
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	data := base64.StdEncoding.EncodeToString(bodyBytes)
	mime := mimetype.Detect(bodyBytes)
	// !! attention !!
	// 对于google的模型，不需要加上data:audio/wav;base64,前缀
	return mime, data, nil
}
