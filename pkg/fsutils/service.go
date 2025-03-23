// The fsutils package provides utilities for managing files in an S3 bucket,
// including uploading, downloading, and deleting objects.
package fsutils

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type FileService interface {
	UploadFile(fileName string, file io.Reader) (string, error)
	DownloadFile(fileName string) ([]byte, error)
	DeleteFile(fileName string) error
	DeleteFiles(fileNames []string) error
}

type Service struct {
	bucket     string
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

// NewService creates a new fileutils Service instance that can be used to upload, download, and delete files from S3.
func NewService(region, bucket, key, secret string) *Service {
	//nolint: exhaustruct
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
	})
	if err != nil {
		panic(fmt.Errorf("failed to create AWS session: %w", err))
	}

	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)
	s3Client := s3.New(sess)

	return &Service{bucket: bucket, client: s3Client, uploader: uploader, downloader: downloader}
}

// UploadFile uploads a file.
func (s *Service) UploadFile(fileName string, file io.Reader) (string, error) {
	//nolint: exhaustruct
	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return result.Location, nil
}

// DownloadFile downloads a file.
/*
handler code:
		// Set headers for the response
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(buf.Bytes())))

		// Write the file to the response
		_, err = w.Write(buf.Bytes())
		if err != nil {
			http.Error(w, "Failed to write file to response", http.StatusInternalServerError)
			return
		}
*/
func (s *Service) DownloadFile(fileName string) ([]byte, error) {
	buf := aws.NewWriteAtBuffer([]byte{})

	//nolint: exhaustruct
	_, err := s.downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return buf.Bytes(), nil
}

// DeleteFile deletes a file from S3.
func (s *Service) DeleteFile(fileName string) error {
	//nolint: exhaustruct
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// / DeleteFiles deletes multiple files from S3.
func (s *Service) DeleteFiles(fileNames []string) error {
	// Convert file names to slice of pointers
	objects := make([]*s3.ObjectIdentifier, 0, len(fileNames))

	for _, fileName := range fileNames {
		//nolint: exhaustruct
		objects = append(objects, &s3.ObjectIdentifier{
			Key: aws.String(fileName),
		})
	}

	//nolint: exhaustruct
	_, err := s.client.DeleteObjects(&s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucket),
		Delete: &s3.Delete{
			Objects: objects,
			Quiet:   aws.Bool(false), // Set to true to suppress errors for individual objects
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete files: %w", err)
	}

	return nil
}
