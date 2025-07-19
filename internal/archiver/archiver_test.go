package archiver

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "archiver_test_dir_")
	if err != nil {
		t.Fatalf("TempDir error: %v", err)
	}
	defer os.RemoveAll(dir)
	testPath := filepath.Join(dir, "subdir1", "subdir2")
	if err := EnsureDir(testPath); err != nil {
		t.Errorf("EnsureDir failed: %v", err)
	}
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("Directory was not created")
	}
}

func TestCreateZip(t *testing.T) {
	dir, err := ioutil.TempDir("", "archiver_test_zip_")
	if err != nil {
		t.Fatalf("TempDir error: %v", err)
	}
	defer os.RemoveAll(dir)
	file1 := filepath.Join(dir, "file1.txt")
	file2 := filepath.Join(dir, "file2.txt")
	_ = ioutil.WriteFile(file1, []byte("hello"), 0644)
	_ = ioutil.WriteFile(file2, []byte("world"), 0644)
	zipPath := filepath.Join(dir, "test.zip")
	if err := CreateZip(zipPath, []string{file1, file2}); err != nil {
		t.Errorf("CreateZip failed: %v", err)
	}
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Errorf("Zip file was not created")
	}
}
