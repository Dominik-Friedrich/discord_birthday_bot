package tempfiles

import (
	"errors"
	log "github.com/chris-dot-exe/AwesomeLog"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const logPrefix = "[TempFileManager]"

var ErrFileNotFound = errors.New("file not found")
var ErrTooManyFiles = errors.New("too many files")
var ErrCleanupRoutineAlreadyStarted = errors.New("cleanup routine has already been started")
var ErrNoFileName = errors.New("fileName is required")

// TempFileManager represents a thread-safe manager for temporary files in a directory.
type TempFileManager struct {
	// directory the files are stored in
	directory string
	// maxFileCount is the maximum amount of files allowed. If this is reached the TempFileManager tries to force clean
	// the oldest entries.
	maxFileCount int
	// forceCleanupCount is the amount of old files cleaned on a force clean
	forceCleanupCount int

	// files is a map with all the downloaded files and the last time they were accesses
	files      map[string]time.Time
	filesMutex sync.Mutex

	started atomic.Bool
	// cleanupCycle is the cycle for the cleanup routine. It triggers every cleanupCycle duration
	cleanupCycle time.Duration
	// fileExpiration is the duration after which files will be cleaned by the cleanup routine
	fileExpiration time.Duration
	close          chan struct{}
}

// NewTempFileManager creates a new TempFileManager with the specified directory and maximum file count.
func NewTempFileManager(
	directory string,
	maxFileCount int,
	forceCleanupCount int,
	cleanupCycle time.Duration,
	fileExpiration time.Duration,

) *TempFileManager {
	m := &TempFileManager{
		directory:         directory,
		maxFileCount:      maxFileCount,
		forceCleanupCount: forceCleanupCount,
		files:             make(map[string]time.Time),
		filesMutex:        sync.Mutex{},
		cleanupCycle:      cleanupCycle,
		fileExpiration:    fileExpiration,
	}

	m.loadInitialFiles()

	return m
}

// Start starts the cleanup routine. Return an ErrCleanupRoutineAlreadyStarted if Start has been called already.
func (tfm *TempFileManager) Start() error {
	if tfm.started.Load() {
		return ErrCleanupRoutineAlreadyStarted
	}

	tfm.started.Store(true)
	go tfm.asyncFileCleanupRoutine()
	return nil
}

// Close closes the cleanup routine permanently. Cannot be started again
func (tfm *TempFileManager) Close() error {
	close(tfm.close)
	return nil
}

// AddFile adds an existing file from the managed directory to the manager.
//
// If the file does not exist an ErrFileNotFound is returned.
// If no fileName is given an ErrNoFileName is returned.
func (tfm *TempFileManager) AddFile(fileName string) error {
	if fileName == "" {
		return ErrNoFileName
	}

	tfm.filesMutex.Lock()
	defer tfm.filesMutex.Unlock()

	// if max files is reached, try to clear the oldest of them
	if len(tfm.files) >= tfm.maxFileCount {
		tfm.cleanOldestFiles(tfm.forceCleanupCount)
	}

	// if still at max return error
	if len(tfm.files) >= tfm.maxFileCount {
		return ErrTooManyFiles
	}

	if !fileExists(tfm.getFilePath(fileName)) {
		return ErrFileNotFound
	}

	tfm.files[fileName] = time.Now()
	log.Printf(log.DEBUG, logPrefix+" added file: %s", fileName)

	return nil
}

// GetFilePath returns filepath for the specified fileName.
// If the file does not exist in the manager an ErrFileNotFound is returned.
// If no fileName is given an ErrNoFileName is returned.
func (tfm *TempFileManager) GetFilePath(fileName string) (string, error) {
	if fileName == "" {
		return "", ErrNoFileName
	}

	tfm.filesMutex.Lock()
	defer tfm.filesMutex.Unlock()

	_, ok := tfm.files[fileName]
	if !ok {
		return "", ErrFileNotFound
	}
	tfm.files[fileName] = time.Now()

	return tfm.getFilePath(fileName), nil
}

func (tfm *TempFileManager) getFilePath(fileName string) string {
	filePath := filepath.Join(tfm.directory, fileName)

	return filePath
}

// cleanOldestFiles cleans the oldest n amount of files from the manager.
//
// It assumes the calling function has already acquired the filesMutex
func (tfm *TempFileManager) cleanOldestFiles(n int) {
	log.Printf(log.DEBUG, logPrefix+" trying to delete %d oldest files", n)

	var files []struct {
		fileName     string
		lastAccessed time.Time
	}

	for f, t := range tfm.files {
		files = append(files, struct {
			fileName     string
			lastAccessed time.Time
		}{fileName: f, lastAccessed: t})
	}

	// sort by time
	sort.Slice(files, func(i, j int) bool {
		return files[i].lastAccessed.Before(files[j].lastAccessed)
	})

	oldestEntries := files[:n]

	deletedCount := 0
	for _, entry := range oldestEntries {
		file := filepath.Join(tfm.directory, entry.fileName)
		err := os.Remove(file)
		if err != nil {
			log.Printf(log.WARN, logPrefix+" error deleting file '%s': %s", file, err)
		}
		delete(tfm.files, entry.fileName)
		deletedCount++
	}

	log.Printf(log.DEBUG, logPrefix+" deleted %d files", deletedCount)
}

func (tfm *TempFileManager) asyncFileCleanupRoutine() {
	ticker := time.NewTicker(tfm.cleanupCycle)
	for {
		select {
		case <-ticker.C:
			tfm.fileCleanup()
		case <-tfm.close:
			return
		}
	}
}

func (tfm *TempFileManager) fileCleanup() {
	log.Printf(log.DEBUG, logPrefix+" running cleanup routine")
	tfm.filesMutex.Lock()

	cutoffTime := time.Now().Add(-tfm.fileExpiration)
	filesToRemove := make([]string, 0)

	for fileName, lastAccessed := range tfm.files {
		if lastAccessed.Before(cutoffTime) {
			filesToRemove = append(filesToRemove, fileName)
		}
	}

	cleaned := 0
	for _, fileName := range filesToRemove {
		file := filepath.Join(tfm.directory, fileName)
		err := os.Remove(file)
		if err != nil {
			log.Printf(log.WARN, logPrefix+" error deleting file '%s': %s", file, err)
			continue
		}
		delete(tfm.files, fileName)
		cleaned++
	}

	log.Printf(log.DEBUG, logPrefix+" cleaned %d files", cleaned)
	tfm.filesMutex.Unlock()
}

// loadInitialFiles loads all the files in the directory.
//
// Ignores subdirectories and any file inside of them
func (tfm *TempFileManager) loadInitialFiles() {
	log.Printf(log.DEBUG, logPrefix+" loading initial files...")

	dir, err := os.ReadDir(tfm.directory)
	if err != nil {
		log.Printf(log.WARN, logPrefix+" could not load initial files: ", err)
	}

	for _, file := range dir {
		if file.IsDir() {
			continue
		}

		err := tfm.AddFile(file.Name())
		if err != nil {
			log.Printf(log.WARN, logPrefix+" could not load initial file: ", err)
			continue
		}
	}

	log.Printf(log.DEBUG, logPrefix+" loaded %d initial files", len(tfm.files))
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
