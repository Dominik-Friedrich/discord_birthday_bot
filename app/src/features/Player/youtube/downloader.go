package youtube

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Kunal-Diwan/go-get-youtube/youtube"
	"io"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
)

const videoIdQueryParam = "v"

var (
	ErrNoId                = errors.New("url has no video id query parameters")
	ErrMalformedQueryParam = errors.New("url has multiple video ids in query parameters")
	ErrVideoTooLong        = errors.New("result video is too long")
)

type Downloader struct {
	maxVideoLength int
}

func (d *Downloader) GetAudio(url string) (interface{}, error) {
	videoId, err := getVideoId(url)
	if err != nil {
		return nil, fmt.Errorf("could not extract videoId, url=%s, err: %s", url, err)
	}

	video, err := youtube.Get(videoId)
	video = video
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (d *Downloader) Query(query string) (*QueryData, error) {
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)

	script := filepath.Join(currentDir, "query.py")

	run := exec.Command("python", script, "-query", query)
	queryOut, err := run.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error creating query command: %s: %s", err, string(queryOut))
	}

	outBuf := bytes.NewBuffer(queryOut)

	jsonQueryData, err := io.ReadAll(outBuf)
	if err != nil {
		return nil, fmt.Errorf("error reading query command output: %s", err)
	}

	var queryData QueryData
	err = json.Unmarshal(jsonQueryData, &queryData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling query command output: %s", err)
	}

	return &queryData, nil
}

func getVideoId(uri string) (string, error) {
	parsedUrl, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	queryParams := parsedUrl.Query()

	videoId, ok := queryParams[videoIdQueryParam]
	if !ok {
		return "", ErrNoId
	}

	if len(videoId) > 1 {
		return "", ErrMalformedQueryParam
	}

	return videoId[0], nil
}
