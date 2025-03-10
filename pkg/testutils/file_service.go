package testutils

import (
	"io"
)

type MockFileService struct {
	UploadedFileName   string
	DownloadedFileName string
	DeletedFileNames   []string
	DownloadedFile     []byte
	UploadedFile       io.Reader
	UploadedLocation   string
}

func NewMockFileService() *MockFileService {
	//nolint: exhaustruct
	return &MockFileService{}
}

func (s *MockFileService) UploadFile(fileName string, file io.Reader) (string, error) {
	s.UploadedFileName = fileName
	s.UploadedFile = file

	return s.UploadedLocation, nil
}

func (s *MockFileService) DownloadFile(fileName string) ([]byte, error) {
	s.DownloadedFileName = fileName

	return s.DownloadedFile, nil
}

func (s *MockFileService) DeleteFile(fileName string) error {
	s.DeletedFileNames = append(s.DeletedFileNames, fileName)

	return nil
}

func (s *MockFileService) DeleteFiles(fileNames []string) error {
	s.DeletedFileNames = append(s.DeletedFileNames, fileNames...)

	return nil
}
