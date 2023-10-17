package tempfiles

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	unused = 0
)

func TestAddFile_ExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	tfm := NewTempFileManager(tempDir, 2, unused, unused, unused)

	// Create a test file
	filePath := filepath.Join(tempDir, "testfile.txt")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// Add the test file to TempFileManager
	err = tfm.AddFile("testfile.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestAddFile_NonExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	tfm := NewTempFileManager(tempDir, 2, unused, unused, unused)

	// Attempt to add a non-existing file
	err := tfm.AddFile("nonexistent.txt")
	if !errors.Is(err, ErrFileNotFound) {
		t.Errorf("Expected ErrFileNotFound, got: %v", err)
	}
}

func TestAddFile_MaxFileCountReached(t *testing.T) {
	tempDir := t.TempDir()
	tfm := NewTempFileManager(tempDir, 2, unused, unused, unused)

	// Create two test files
	filePath1 := filepath.Join(tempDir, "testfile1.txt")
	f, err := os.Create(filePath1)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	filePath2 := filepath.Join(tempDir, "testfile2.txt")
	f, err = os.Create(filePath2)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	filePath3 := filepath.Join(tempDir, "testfile3.txt")
	f, err = os.Create(filePath2)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Add the test files to TempFileManager
	err = tfm.AddFile("testfile1.txt")
	if err != nil {
		t.Fatal(err)
	}
	err = tfm.AddFile("testfile2.txt")
	if err != nil {
		t.Fatal(err)
	}

	err = tfm.AddFile(filePath3)
	if !errors.Is(err, ErrTooManyFiles) {
		t.Errorf("Expected ErrTooManyFiles, got: %v", err)
	}
}

func TestAddFile_CleanOldest(t *testing.T) {
	tempDir := t.TempDir()
	tfm := NewTempFileManager(tempDir, 4, 2, unused, unused)
	tmpFileCount := 4

	for i := 1; i <= tmpFileCount; i++ {
		fileName := fmt.Sprintf("testfile%d.txt", i)
		filePath := filepath.Join(tempDir, fileName)
		f, err := os.Create(filePath)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()

		err = tfm.AddFile(fileName)
		if err != nil {
			t.Fatal(err)
		}

		// sleep to get viable oldest values
		time.Sleep(time.Millisecond * 10)
	}

	fileName := "testfile5.txt"
	filePath := filepath.Join(tempDir, fileName)
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	err = tfm.AddFile(fileName)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, tfm.files, 3, "Expected cleanup to have removed 2 files before adding a third again")

	_, ok := tfm.files[fileName]
	assert.True(t, ok, "Expected new entry to be added to map")

	_, ok = tfm.files["testfile1d.txt"]
	assert.False(t, ok, "Expected oldest entry to be removed from map")

	_, ok = tfm.files["testfile2d.txt"]
	assert.False(t, ok, "Expected new entry to be removed map")
}

func TestCleanupRoutine_RemovesOnlyExpired(t *testing.T) {
	tempDir := t.TempDir()
	tfm := NewTempFileManager(tempDir, 4, unused, time.Second, time.Second)

	filePath1 := filepath.Join(tempDir, "testfile1.txt")
	f, err := os.Create(filePath1)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	filePath2 := filepath.Join(tempDir, "testfile2.txt")
	f, err = os.Create(filePath2)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	err = tfm.AddFile("testfile1.txt")
	if err != nil {
		t.Fatal(err)
	}

	err = tfm.Start()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 500)

	err = tfm.AddFile("testfile2.txt")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 800)

	assert.Len(t, tfm.files, 1, "Expected cleanup to have removed 1 file")

	_, ok := tfm.files["testfile1.txt"]
	assert.False(t, ok, "Expected old entry to be removed map")

	_, ok = tfm.files["testfile2.txt"]
	assert.True(t, ok, "Expected new entry to stay in map")
}

func TestGetFilePath_ExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	tfm := NewTempFileManager(tempDir, 2, unused, unused, unused)

	// Create a test file
	filePath := filepath.Join(tempDir, "testfile.txt")
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Add the test file to TempFileManager
	err = tfm.AddFile("testfile.txt")
	if err != nil {
		t.Fatal(err)
	}

	lastAccessedBefore := tfm.files["testfile.txt"]

	time.Sleep(time.Second)

	// Get the file path for an existing file
	fp, err := tfm.GetFilePath("testfile.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if fp != filePath {
		t.Errorf("Expected file path to be %s, got: %s", filePath, fp)
	}

	// Verify that the last accessed time was updated
	lastAccessed := tfm.files["testfile.txt"]
	if lastAccessedBefore.Equal(lastAccessed) {
		t.Errorf("Expected lastAccessed time to be updated, got: %s", lastAccessed)
	}

	if !lastAccessedBefore.Before(lastAccessed) {
		t.Errorf("Expected lastAccessed time to be newer than before, got: %s", lastAccessed)
	}
}

func TestGetFilePath_NonExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	tfm := NewTempFileManager(tempDir, 2, unused, unused, unused)

	// Attempt to get the file path for a non-existing file
	_, err := tfm.GetFilePath("nonexistent.txt")
	if !errors.Is(err, ErrFileNotFound) {
		t.Errorf("Expected ErrFileNotFound, got: %v", err)
	}
}

func TestStart_AlreadyStarted(t *testing.T) {
	tempDir := t.TempDir()
	tfm := NewTempFileManager(tempDir, 2, unused, time.Second, time.Second)

	err := tfm.Start()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	err = tfm.Start()
	if !errors.Is(err, ErrCleanupRoutineAlreadyStarted) {
		t.Errorf("Expected ErrCleanupRoutineAlreadyStarted, got: %v", err)
	}
}
