package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/dorklyorg/dorkly/internal/dorkly"
	"os"
)

const (
	dorklyYamlEnvVar           = "DORKLY_YAML"
	defaultDorklyYamlInputPath = "project"

	s3BucketEnvVar = "DORKLY_S3_BUCKET"
)

var logger = dorkly.GetLogger()

func main() {
	ctx := context.Background()
	dorklyYamlInputPath := os.Getenv(dorklyYamlEnvVar)
	if dorklyYamlInputPath == "" {
		logger.Infof("Env var [%s] not set. Using default: %s", dorklyYamlEnvVar, defaultDorklyYamlInputPath)
		dorklyYamlInputPath = defaultDorklyYamlInputPath
	}

	s3Bucket := os.Getenv(s3BucketEnvVar)
	if s3Bucket == "" {
		logger.Fatalf("Required env var [%s] not set.", s3BucketEnvVar)
	}

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Fatalf("Couldn't load default aws configuration. Have you set up your AWS account? %v", err)
		return
	}

	logAwsCallerIdentity(awsConfig, ctx)

	s3Client := s3.NewFromConfig(awsConfig)
	s3ArchiveService, err := dorkly.NewS3RelayArchiveService(s3Client, s3Bucket)
	if err != nil {
		logger.Fatal(err)
	}
	reconciler := dorkly.NewReconciler(s3ArchiveService, dorklyYamlInputPath)

	err = reconciler.Reconcile(context.Background())
	if err != nil {
		logger.Fatal(err)
	}
}

func logAwsCallerIdentity(awsConfig aws.Config, ctx context.Context) {
	svc := sts.NewFromConfig(awsConfig)
	input := &sts.GetCallerIdentityInput{}

	result, err := svc.GetCallerIdentity(ctx, input)
	if err != nil {
		logger.Fatal(err)
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("AWS Identity: %v", string(jsonBytes))
}
