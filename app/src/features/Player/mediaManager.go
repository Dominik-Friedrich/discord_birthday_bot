package Player

import (
	"fmt"
	log "github.com/chris-dot-exe/AwesomeLog"
	"main/src/features/Player/youtube"
	"main/src/lib/tempfiles"
	"path"
	"path/filepath"
	"time"
)

const logPrefix = "[MediaManager] "

type MediaManager struct {
	directory      string
	maxFiles       int
	maxVideoLength time.Duration

	fileManager *tempfiles.TempFileManager
}

func NewMediaManager(directory string, maxFileCount int, maxVideoLength time.Duration) *MediaManager {
	m := new(MediaManager)
	m.fileManager = tempfiles.NewTempFileManager(
		filepath.Dir("./resources"),
		maxFileCount,
		maxFileCount/10,
		5*time.Minute,
		10*time.Minute,
	)
	err := m.fileManager.Start()
	if err != nil {
		log.Panicf("error starting fileManager cleanup routine")
	}

	return m
}

func (m *MediaManager) GetMediaFilePathByQuery(query string) (string, error) {
	// query youtube to get video id
	mediaInfo, err := youtube.Query(query)
	if err != nil {
		return "", fmt.Errorf("error getting media info: %s", err)
	}
	if mediaInfo.Error != "" {
		return "", fmt.Errorf("error getting media info: %s", mediaInfo.Error)
	}

	filePath, err := m.GetMediaFilePathByFileName(mediaInfo.VideoInfo.Filename)
	if err == nil {
		return filePath, nil
	}

	// file not present in file manager -> attempt to download media
	log.Printf(log.INFO, logPrefix+"could not get media: %s", err)
	log.Printf(log.INFO, logPrefix+"attempting to download it")
	fileName, err := m.downloadMedia(query)
	log.Printf(log.INFO, logPrefix+"media downloaded to %s", path.Join(m.directory, fileName))

	err = m.fileManager.AddFile(fileName)
	if err != nil {
		return "", fmt.Errorf("could not add media to manager: %s", err)
	}

	return fileName, nil
}

func (m *MediaManager) GetMediaFilePathByFileName(fileName string) (string, error) {
	return m.fileManager.GetFilePath(fileName)
}

func (m *MediaManager) downloadMedia(query string) (string, error) {
	mediaInfo, err := youtube.Download(query, m.maxVideoLength, m.directory)
	if err != nil {
		return "", fmt.Errorf("error downloading media: %s", err)
	}
	if mediaInfo.Error != "" {
		return "", fmt.Errorf("error downloading media info: %s", mediaInfo.Error)
	}

	return mediaInfo.VideoInfo.Filename, nil
}
