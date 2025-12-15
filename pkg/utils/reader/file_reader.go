package reader

import (
	"os"
)

type FileReader struct{}

func NewFileReader() *FileReader {
	return &FileReader{}
}

func (fr *FileReader) ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
