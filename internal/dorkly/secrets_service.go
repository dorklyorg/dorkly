package dorkly

import "context"

type SecretsService interface {
	GetSdkKey(ctx context.Context, project, env string) (string, error)
	GetMobileKey(ctx context.Context, project, env string) (string, error)
}

type awsSecretsService struct {
}

// Secrets naming convention must be kept in sync with the terraform bits here:
// https://github.com/dorklyorg/terraform-aws-dorkly-flags/blob/main/dorkly_environment/main.tf
func sdkKeySecretName(project, env string) string {
	return "dorkly-" + project + "-" + env + "-sdk-key"
}

// Secrets naming convention must be kept in sync with the terraform bits here:
// https://github.com/dorklyorg/terraform-aws-dorkly-flags/blob/main/dorkly_environment/main.tf
func mobileKeySecretName(project, env string) string {
	return "dorkly-" + project + "-" + env + "-mob-key"
}
