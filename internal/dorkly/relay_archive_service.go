package dorkly

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	s3ObjectKey = "flags.tar.gz"
)

var (
	ErrExistingArchiveNotFound = fmt.Errorf("existing relay archive not found")
)

type RelayArchiveService interface {
	fetchExisting(ctx context.Context) (*RelayArchive, error)
	saveNew(ctx context.Context, relayArchive RelayArchive) error
}

var _ RelayArchiveService = &S3RelayArchiveService{}

type S3RelayArchiveService struct {
	bucket   string
	s3Client *s3.Client
}

func NewS3RelayArchiveService(s3Client *s3.Client, bucket string) (RelayArchiveService, error) {
	return S3RelayArchiveService{
		bucket:   bucket,
		s3Client: s3Client,
	}, nil
}

func (s S3RelayArchiveService) fetchExisting(ctx context.Context) (*RelayArchive, error) {
	existingRelayArchiveFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("dorkly-%v.tar.gz", time.Now().UnixMicro()))

	log.Printf("Fetching existing relay archive from S3 bucket: [%s] with object key: [%s]. Saving to [%s]",
		s.bucket, s3ObjectKey, existingRelayArchiveFilePath)

	goo, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3ObjectKey),
	})

	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, ErrExistingArchiveNotFound
		}
		return nil, err
	}

	outFile, err := os.Create(existingRelayArchiveFilePath)
	if err != nil {
		return nil, err
	}
	defer outFile.Close()
	defer goo.Body.Close()

	_, err = io.Copy(outFile, goo.Body)
	if err != nil {
		return nil, err
	}
	existing, err := loadRelayArchiveFromTarGzFile(existingRelayArchiveFilePath)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (s S3RelayArchiveService) saveNew(ctx context.Context, relayArchive RelayArchive) error {
	archiveFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("dorkly-%v.tar.gz", time.Now().UnixMicro()))

	log.Printf("Uploading new relay archive to S3 bucket: [%s] with object key: [%s]. Also saving to [%s]",
		s.bucket, s3ObjectKey, archiveFilePath)

	err := relayArchive.toTarGzFile(archiveFilePath)
	if err != nil {
		return err
	}

	file, err := os.Open(archiveFilePath)
	if err != nil {
		return err
	}

	_, err = s.s3Client.PutObject(ctx,
		&s3.PutObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(s3ObjectKey),
			Body:   file,
		},
	)

	return err
}

var _ RelayArchiveService = &LocalFileRelayArchiveService{}

type LocalFileRelayArchiveService struct {
	archivePath string
}

// NewLocalFileRelayArchiveService creates a RelayArchiveService that reads and writes to a local file. It is intended for testing.
// It uses one path for both reading and writing which means it will overwrite an existing archive when calling saveNew
func NewLocalFileRelayArchiveService(archivePath string) RelayArchiveService {
	return &LocalFileRelayArchiveService{
		archivePath: archivePath,
	}
}

func (l *LocalFileRelayArchiveService) fetchExisting(ctx context.Context) (*RelayArchive, error) {
	existing, err := loadRelayArchiveFromTarGzFile(l.archivePath)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (l *LocalFileRelayArchiveService) saveNew(ctx context.Context, relayArchive RelayArchive) error {
	return relayArchive.toTarGzFile(l.archivePath)
}
