package storage

import (
	"log"
	"os"
)

type FileStorage struct {
	directory string
}

func NewFileStorage(directory string) *FileStorage {
	err := os.Chdir(directory)
	if err != nil {
		log.Fatalln("Failed open directory for file storage")
		return nil
	}
	return &FileStorage{directory: directory}
}
