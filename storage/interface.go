package storage

import "time"

type Storage interface {
	UploadObject(data []byte, path string) (string, error)
	DeleteObject(string) error
	GetFullUrl(path string) string
	ListAll() ([]string, error)
	GeneratePresignedURL(objectPath string, expiration time.Duration) (string, error)
}
