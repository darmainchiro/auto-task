package external

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileStorage struct {
	uploadPath string
}

func NewFileStorage() *FileStorage {
	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath == "" {
		uploadPath = "./uploads"
	}

	// Create upload directory if not exists
	os.MkdirAll(uploadPath, 0755)

	return &FileStorage{uploadPath: uploadPath}
}

func (fs *FileStorage) SaveFile(filename string, data []byte) (string, error) {
	// Generate unique filename
	timestamp := time.Now().Unix()
	uniqueFilename := fmt.Sprintf("%d_%s", timestamp, filename)
	filePath := filepath.Join(fs.uploadPath, uniqueFilename)

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	return filePath, nil
}

func (fs *FileStorage) GetFile(filepath string) ([]byte, error) {
	return os.ReadFile(filepath)
}

func (fs *FileStorage) DeleteFile(filepath string) error {
	return os.Remove(filepath)
}
