package youtube

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"main/src/lib/types"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
)

var (
	ErrNoId                = errors.New("url has no video id query parameters")
	ErrMalformedQueryParam = errors.New("url has multiple video ids in query parameters")
)

func Query(query string) (*QueryData, error) {
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

func Download(query string, maxLength types.Second, destDir string) (*QueryData, error) {
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)

	script := filepath.Join(currentDir, "download.py")

	run := exec.Command("python", script, "-query", query, "-max_duration", fmt.Sprintf("%d", maxLength))
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
