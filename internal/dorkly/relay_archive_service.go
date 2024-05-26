package dorkly

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	s3ObjectKey = "flags.tar.gz"
)

var (
	ErrExistingArchiveNotFound = errors.New("existing relay archive not found")
)

type RelayArchiveService interface {
	fmt.Stringer
	fetchExisting(ctx context.Context) (*RelayArchive, error)
	saveNew(ctx context.Context, relayArchive RelayArchive) error
}

var _ RelayArchiveService = &s3RelayArchiveService{}

type s3RelayArchiveService struct {
	bucket   string
	s3Client *s3.Client
	logger   *zap.SugaredLogger
}

func (s s3RelayArchiveService) String() string {
	return fmt.Sprintf("s3RelayArchiveService using bucket: [%s]", s.bucket)
}

func NewS3RelayArchiveService(s3Client *s3.Client, bucket string) (RelayArchiveService, error) {
	return s3RelayArchiveService{
		bucket:   bucket,
		s3Client: s3Client,
		logger: logger.Named("s3RelayArchiveService").
			With(zap.String("bucket", bucket)).
			With(zap.String("objectKey", s3ObjectKey)),
	}, nil
}

func (s s3RelayArchiveService) fetchExisting(ctx context.Context) (*RelayArchive, error) {
	existingRelayArchiveFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("dorkly-%v.tar.gz", time.Now().UnixMicro()))

	s.logger.Infof("Fetching existing relay archive from S3. Saving to [%s]", existingRelayArchiveFilePath)

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
	defer goo.Body.Close()

	outFile, err := os.Create(existingRelayArchiveFilePath)
	if err != nil {
		return nil, err
	}
	defer outFile.Close()

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

func (s s3RelayArchiveService) saveNew(ctx context.Context, relayArchive RelayArchive) error {
	archiveFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("dorkly-%v.tar.gz", time.Now().UnixMicro()))
	s.logger.Infof("Uploading new relay archive to S3. Also saving to [%s]", archiveFilePath)

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

var _ RelayArchiveService = &localFileRelayArchiveService{}

type localFileRelayArchiveService struct {
	archivePath string
}

func (l *localFileRelayArchiveService) String() string {
	return fmt.Sprintf("localFileRelayArchiveService using path: [%s]", l.archivePath)
}

// NewLocalFileRelayArchiveService creates a RelayArchiveService that reads and writes to a local file. It is intended for testing.
// It uses one path for both reading and writing which means it will overwrite an existing archive when calling saveNew
func NewLocalFileRelayArchiveService(archivePath string) RelayArchiveService {
	return &localFileRelayArchiveService{
		archivePath: archivePath,
	}
}

func (l *localFileRelayArchiveService) fetchExisting(ctx context.Context) (*RelayArchive, error) {
	existing, err := loadRelayArchiveFromTarGzFile(l.archivePath)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (l *localFileRelayArchiveService) saveNew(ctx context.Context, relayArchive RelayArchive) error {
	return relayArchive.toTarGzFile(l.archivePath)
}
