package youtube

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	ErrNoId                = errors.New("url has no video id query parameters")
	ErrMalformedQueryParam = errors.New("url has multiple video ids in query parameters")
	ErrNotYoutubeUrl       = errors.New("url is not a youtube url")
)

func Query(query string) (*QueryData, error) {
	id, err := getVideoId(query)
	if err == nil {
		query = id
	}

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

func Download(query string, maxLength time.Duration, destDir string) (*QueryData, error) {
	id, err := getVideoId(query)
	if err == nil {
		query = id
	}

	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)

	script := filepath.Join(currentDir, "download.py")

	run := exec.Command("python", script, "-query", query, "-max_duration", fmt.Sprintf("%d", int(maxLength.Seconds())))
	run.Dir = destDir
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

	if !strings.Contains(parsedUrl.Host, "youtube") {
		return "", ErrNotYoutubeUrl
	}

	if strings.Contains(parsedUrl.Path, "shorts") {
		return getVideoIdShortsUrl(parsedUrl)
	}

	queryParams := parsedUrl.Query()

	const videoIdQueryParam = "v"
	videoId, ok := queryParams[videoIdQueryParam]
	if !ok {
		return "", ErrNoId
	}

	if len(videoId) > 1 {
		return "", ErrMalformedQueryParam
	}

	return videoId[0], nil
}

func getVideoIdShortsUrl(shortsUri *url.URL) (string, error) {
	path := shortsUri.Path

	pathElems := strings.Split(path, "/")

	if len(pathElems) != 3 {
		return "", ErrNoId
	}

	return pathElems[2], nil
}
