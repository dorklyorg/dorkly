package dorkly

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"path/filepath"
	"testing"
)

const (
	awsRegion = "us-west-2"
)

func Test_S3RelayArchiveService(t *testing.T) {
	ctx := context.Background()

	localstackContainer, err := localstack.Run(ctx,
		"localstack/localstack:3.4",
		testcontainers.WithEnv(map[string]string{"SERVICES": "s3"}),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := localstackContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	s3Client, err := s3Client(ctx, localstackContainer)
	require.Nil(t, err)

	// Create a bucket
	bucketName := "test-bucket"
	_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: &bucketName,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: awsRegion,
		}})
	require.Nil(t, err)

	archiveService, err := NewS3RelayArchiveService(s3Client, bucketName)
	require.Nil(t, err)

	testProject1Archive := testProject1.toRelayArchive()

	err = archiveService.saveNew(ctx, *testProject1Archive)
	require.Nil(t, err)

	existingArchive, err := archiveService.fetchExisting(ctx)
	require.Nil(t, err)

	require.Equal(t, testProject1Archive, existingArchive)
}

func s3Client(ctx context.Context, l *localstack.LocalStackContainer) (*s3.Client, error) {
	mappedPort, err := l.MappedPort(ctx, "4566/tcp")
	if err != nil {
		return nil, err
	}

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return nil, err
	}
	defer provider.Close()

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		return nil, err
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("accesskey", "secret", "token")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg,
		func(o *s3.Options) {
			o.UsePathStyle = true
			o.BaseEndpoint = aws.String(fmt.Sprintf("http://%s:%d", host, mappedPort.Int()))
		},
	)

	return client, nil
}

func Test_LocalFileRelayArchiveService(t *testing.T) {
	ctx := context.Background()
	var err error

	archivePath := filepath.Join(t.TempDir(), "dorkly-existing.tar.gz")
	archiveService := NewLocalFileRelayArchiveService(archivePath)

	testProject1Archive := testProject1.toRelayArchive()
	t.Run("saveNew", func(t *testing.T) {
		err = archiveService.saveNew(ctx, *testProject1Archive)
		require.Nil(t, err)
	})

	var existingArchive *RelayArchive
	t.Run("fetchExisting", func(t *testing.T) {
		existingArchive, err = archiveService.fetchExisting(ctx)
		require.Nil(t, err)
	})

	assert.Equal(t, testProject1Archive, existingArchive)
}
