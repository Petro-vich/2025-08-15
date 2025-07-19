package archiver

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

func CreateZip(zipPath string, files []string) error {
	if err := EnsureDir(filepath.Dir(zipPath)); err != nil {
		return fmt.Errorf("failed to create archive dir: %w", err)
	}
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	successLoad := 0
	for _, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			continue
		}
		w, err := zipWriter.Create(filepath.Base(filePath))
		if err != nil {
			file.Close()
			continue
		}
		_, err = io.Copy(w, file)
		file.Close()
		if err != nil {
			continue
		}
		successLoad++
	}
	if successLoad == 0 {
		return fmt.Errorf("no files were added to the archive")
	}
	return nil
}
