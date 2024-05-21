package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/dorklyorg/dorkly/internal/dorkly"
	"log"
	"os"
)

const (
	dorklyYamlEnvVar           = "DORKLY_YAML"
	defaultDorklyYamlInputPath = "project"

	s3BucketEnvVar = "DORKLY_S3_BUCKET"
)

func main() {
	ctx := context.Background()
	dorklyYamlInputPath := os.Getenv(dorklyYamlEnvVar)
	if dorklyYamlInputPath == "" {
		log.Printf("Env var [%s] not set. Using default: %s", dorklyYamlEnvVar, defaultDorklyYamlInputPath)
		dorklyYamlInputPath = defaultDorklyYamlInputPath
	}

	s3Bucket := os.Getenv(s3BucketEnvVar)
	if s3Bucket == "" {
		log.Fatalf("Required env var [%s] not set.", s3BucketEnvVar)
	}

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Couldn't load default aws configuration. Have you set up your AWS account? %v", err)
		return
	}

	svc := sts.NewFromConfig(awsConfig)
	input := &sts.GetCallerIdentityInput{}

	result, err := svc.GetCallerIdentity(ctx, input)
	if err != nil {
		log.Fatal(err)
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("AWS Identity: %v", string(jsonBytes))

	s3Client := s3.NewFromConfig(awsConfig)
	s3ArchiveService, err := dorkly.NewS3RelayArchiveService(s3Client, s3Bucket)
	reconciler := dorkly.NewReconciler(s3ArchiveService, dorklyYamlInputPath)

	err = reconciler.Reconcile(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
